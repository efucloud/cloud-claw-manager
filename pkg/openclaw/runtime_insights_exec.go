package openclaw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/cloud-claw-manager/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	openclawConfigInPodPath = "/home/node/.openclaw/openclaw.json"
	openclawCronJobsPath    = "/home/node/.openclaw/cron/jobs.json"
)

var runtimeInsightsCache = struct {
	mu   sync.RWMutex
	data map[string]OpenClawRuntimeInsights
}{
	data: map[string]OpenClawRuntimeInsights{},
}

func RefreshRuntimeInsightsCache(ctx context.Context) error {
	kc, _, _, err := defaultOpenClawClients.get(ctx)
	if err != nil {
		return err
	}
	restCfg, err := kube.BuildRestConfig(ctx, config.RunKubeConfig)
	if err != nil {
		return err
	}
	deployList, err := kc.AppsV1().Deployments("").List(ctx, metav1.ListOptions{
		LabelSelector: OpenClawInstanceLabelInstance,
	})
	if err != nil {
		return err
	}

	type instanceRef struct {
		namespace string
		name      string
	}
	refs := map[string]instanceRef{}
	for i := range deployList.Items {
		dep := deployList.Items[i]
		instanceName := strings.TrimSpace(dep.Labels[OpenClawInstanceLabelInstance])
		if instanceName == "" {
			instanceName = strings.TrimSpace(dep.Annotations[openClawInstanceAnnotation])
		}
		if instanceName == "" {
			instanceName = strings.TrimSpace(dep.Name)
		}
		if strings.TrimSpace(dep.Namespace) == "" || instanceName == "" {
			continue
		}
		key := runtimeInsightKey(dep.Namespace, instanceName)
		refs[key] = instanceRef{namespace: dep.Namespace, name: instanceName}
	}

	keys := make([]string, 0, len(refs))
	for k := range refs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	next := make(map[string]OpenClawRuntimeInsights, len(keys))
	for _, key := range keys {
		ref := refs[key]
		collected, collectErr := collectRuntimeInsightsForInstance(ctx, kc, restCfg, ref.namespace, ref.name)
		if collectErr != nil {
			collected.CollectedAt = time.Now().UTC().Format(time.RFC3339)
			collected.LastError = collectErr.Error()
		}
		next[key] = collected
	}

	runtimeInsightsCache.mu.Lock()
	runtimeInsightsCache.data = next
	runtimeInsightsCache.mu.Unlock()
	return nil
}

func GetRuntimeInsights(namespace, instanceName string) (OpenClawRuntimeInsights, bool) {
	key := runtimeInsightKey(namespace, instanceName)
	runtimeInsightsCache.mu.RLock()
	defer runtimeInsightsCache.mu.RUnlock()
	value, ok := runtimeInsightsCache.data[key]
	return value, ok
}

func runtimeInsightKey(namespace, instanceName string) string {
	return strings.TrimSpace(namespace) + "/" + strings.TrimSpace(instanceName)
}

func collectRuntimeInsightsForInstance(
	ctx context.Context,
	kc *kubernetes.Clientset,
	restCfg *rest.Config,
	namespace, instanceName string,
) (OpenClawRuntimeInsights, error) {
	pods, err := kc.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", OpenClawInstanceLabelInstance, instanceName),
	})
	if err != nil {
		return OpenClawRuntimeInsights{}, err
	}
	pod, container, err := pickTelemetryTargetPod(pods.Items)
	if err != nil {
		return OpenClawRuntimeInsights{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	out := OpenClawRuntimeInsights{
		CollectedAt:     now,
		SourcePod:       pod.Name,
		SourceContainer: container,
		StatusState:     "unknown",
		GatewayState:    "unknown",
		SecurityState:   "unknown",
	}

	statusRaw, statusErr := execShellInPod(ctx, kc, restCfg, namespace, pod.Name, container, "openclaw status --json")
	if statusErr != nil {
		out.StatusState = "error"
		appendInsightError(&out, "status", statusErr)
	} else {
		out.StatusState = "ok"
		var statusRoot map[string]interface{}
		if err := json.Unmarshal([]byte(statusRaw), &statusRoot); err == nil {
			if runtimeVersion, _ := asString(statusRoot["runtimeVersion"]); runtimeVersion != "" {
				out.RuntimeVersion = runtimeVersion
			}
			populateStatusSessionSummary(&out, statusRoot)
		}
	}

	gatewayRaw, gatewayErr := execShellInPod(ctx, kc, restCfg, namespace, pod.Name, container, "openclaw gateway status --json")
	if gatewayErr != nil {
		out.GatewayState = "error"
		appendInsightError(&out, "gateway", gatewayErr)
	} else {
		out.GatewayState = "ok"
		var gatewayRoot map[string]interface{}
		if err := json.Unmarshal([]byte(gatewayRaw), &gatewayRoot); err == nil {
			out.GatewayState = summarizeGatewayState(gatewayRoot)
		}
	}

	securityRaw, securityErr := execShellInPod(ctx, kc, restCfg, namespace, pod.Name, container, "openclaw security audit --json")
	if securityErr != nil {
		out.SecurityState = "error"
		appendInsightError(&out, "security", securityErr)
	} else {
		out.SecurityState = "ok"
		var securityRoot map[string]interface{}
		if err := json.Unmarshal([]byte(securityRaw), &securityRoot); err == nil {
			out.SecurityState = summarizeSecurityState(securityRoot)
		}
	}

	configRaw, configErr := execShellInPod(ctx, kc, restCfg, namespace, pod.Name, container, fmt.Sprintf("cat %s", openclawConfigInPodPath))
	if configErr != nil {
		appendInsightError(&out, "openclaw.json", configErr)
	} else {
		populateConfigSummary(&out, configRaw)
	}

	cronRaw, cronErr := execShellInPod(ctx, kc, restCfg, namespace, pod.Name, container, fmt.Sprintf("cat %s", openclawCronJobsPath))
	if cronErr != nil {
		appendInsightError(&out, "cron/jobs.json", cronErr)
	} else {
		out.CronJobsCount = parseCronJobsCount(cronRaw)
	}

	sessionIndexCountRaw, sessionIndexErr := execShellInPod(
		ctx,
		kc,
		restCfg,
		namespace,
		pod.Name,
		container,
		"ls -1 /home/node/.openclaw/agents/*/sessions/sessions.json 2>/dev/null | wc -l",
	)
	if sessionIndexErr != nil {
		appendInsightError(&out, "agents/*/sessions/sessions.json", sessionIndexErr)
	} else {
		out.SessionIndexFilesCount = parseIntLoose(sessionIndexCountRaw)
	}

	return out, nil
}

func pickTelemetryTargetPod(pods []corev1.Pod) (corev1.Pod, string, error) {
	if len(pods) == 0 {
		return corev1.Pod{}, "", fmt.Errorf("no pod found for instance")
	}
	readyRunning := make([]corev1.Pod, 0, len(pods))
	running := make([]corev1.Pod, 0, len(pods))
	for i := range pods {
		pod := pods[i]
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}
		running = append(running, pod)
		if isPodReady(pod) {
			readyRunning = append(readyRunning, pod)
		}
	}
	candidates := readyRunning
	if len(candidates) == 0 {
		candidates = running
	}
	if len(candidates) == 0 {
		return corev1.Pod{}, "", fmt.Errorf("no running pod found for instance")
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].CreationTimestamp.Time.After(candidates[j].CreationTimestamp.Time)
	})
	target := candidates[0]
	container := pickTelemetryContainer(target.Spec.Containers)
	if container == "" {
		return corev1.Pod{}, "", fmt.Errorf("no container found in pod %s", target.Name)
	}
	return target, container, nil
}

func pickTelemetryContainer(containers []corev1.Container) string {
	if len(containers) == 0 {
		return ""
	}
	for i := range containers {
		name := strings.TrimSpace(containers[i].Name)
		if strings.Contains(strings.ToLower(name), "openclaw") {
			return name
		}
		image := strings.TrimSpace(containers[i].Image)
		if strings.Contains(strings.ToLower(image), "openclaw") {
			return name
		}
	}
	return strings.TrimSpace(containers[0].Name)
}

func isPodReady(pod corev1.Pod) bool {
	for i := range pod.Status.Conditions {
		cond := pod.Status.Conditions[i]
		if cond.Type == corev1.PodReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

func execShellInPod(
	ctx context.Context,
	kc *kubernetes.Clientset,
	restCfg *rest.Config,
	namespace, podName, container, script string,
) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	req := kc.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   []string{"sh", "-lc", script},
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
	if err != nil {
		return "", err
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := exec.StreamWithContext(timeoutCtx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	}); err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func appendInsightError(out *OpenClawRuntimeInsights, source string, err error) {
	if err == nil {
		return
	}
	segment := source + "=" + err.Error()
	if strings.TrimSpace(out.LastError) == "" {
		out.LastError = segment
		return
	}
	out.LastError = out.LastError + "; " + segment
}

func summarizeGatewayState(root map[string]interface{}) string {
	rpc := asMap(root["rpc"])
	if asBool(rpc["ok"]) {
		return "ok"
	}
	service := asMap(root["service"])
	runtime := asMap(service["runtime"])
	state, _ := asString(runtime["status"])
	if strings.TrimSpace(state) == "" {
		state, _ = asString(runtime["state"])
	}
	state = strings.TrimSpace(strings.ToLower(state))
	switch state {
	case "running", "active":
		return "ok"
	case "stopped", "error", "failed":
		return "error"
	default:
		if state == "" {
			return "warn"
		}
		return state
	}
}

func summarizeSecurityState(root map[string]interface{}) string {
	summary := asMap(root["summary"])
	critical := asInt(summary["critical"])
	warn := asInt(summary["warn"])
	if critical > 0 {
		return "critical"
	}
	if warn > 0 {
		return "warn"
	}
	return "ok"
}

func populateStatusSessionSummary(out *OpenClawRuntimeInsights, root map[string]interface{}) {
	if out == nil {
		return
	}
	sessions := asMap(root["sessions"])
	out.SessionCount = asInt(sessions["count"])
	if out.SessionCount == 0 {
		agents := asMap(root["agents"])
		out.SessionCount = asInt(agents["totalSessions"])
	}
	recent, _ := sessions["recent"].([]interface{})
	var latestMs int64
	for i := range recent {
		item := asMap(recent[i])
		updatedAtMs := asInt64(item["updatedAt"])
		if updatedAtMs > latestMs {
			latestMs = updatedAtMs
		}
	}
	if latestMs > 0 {
		out.SessionLastUpdatedAt = time.UnixMilli(latestMs).UTC().Format(time.RFC3339)
	}
}

func populateConfigSummary(out *OpenClawRuntimeInsights, raw string) {
	out.OpenClawConfigPath = openclawConfigInPodPath
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		appendInsightError(out, "openclaw.json.parse", err)
		return
	}

	models := asMap(root["models"])
	out.OpenClawPrimaryModel = firstNonEmptyString(
		asStringMust(models["primary"]),
	)
	providers := models["providers"]
	switch typed := providers.(type) {
	case []interface{}:
		out.OpenClawProvidersCount = len(typed)
	case map[string]interface{}:
		out.OpenClawProvidersCount = len(typed)
	}

	agents := asMap(root["agents"])
	defaults := asMap(agents["defaults"])
	out.OpenClawWorkspace = firstNonEmptyString(
		asStringMust(defaults["workspace"]),
	)
	if out.OpenClawPrimaryModel == "" {
		modelCfg := asMap(defaults["model"])
		out.OpenClawPrimaryModel = asStringMust(modelCfg["primary"])
	}
	list, _ := agents["list"].([]interface{})
	out.OpenClawAgentsCount = len(list)

	gateway := asMap(root["gateway"])
	controlUi := asMap(gateway["controlUi"])
	if value, ok := controlUi["dangerouslyDisableDeviceAuth"]; ok {
		boolValue := asBool(value)
		out.ControlUiDisableDeviceAuth = &boolValue
	}
	allowedOrigins := asStringSlice(controlUi["allowedOrigins"])
	if len(allowedOrigins) > 0 {
		out.ControlUiAllowedOrigins = allowedOrigins
	}
}

func parseCronJobsCount(raw string) int {
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		return 0
	}
	jobs, _ := root["jobs"].([]interface{})
	return len(jobs)
}

func parseIntLoose(raw string) int {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0
	}
	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0
	}
	return value
}

func asMap(value interface{}) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	casted, ok := value.(map[string]interface{})
	if !ok || casted == nil {
		return map[string]interface{}{}
	}
	return casted
}

func asString(value interface{}) (string, bool) {
	if value == nil {
		return "", false
	}
	switch typed := value.(type) {
	case string:
		return typed, true
	default:
		return "", false
	}
}

func asStringMust(value interface{}) string {
	text, _ := asString(value)
	return strings.TrimSpace(text)
}

func asStringSlice(value interface{}) []string {
	out := []string{}
	items, ok := value.([]interface{})
	if !ok {
		return out
	}
	for i := range items {
		text := strings.TrimSpace(asStringMust(items[i]))
		if text == "" {
			continue
		}
		out = append(out, text)
	}
	return out
}

func asBool(value interface{}) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		lower := strings.ToLower(strings.TrimSpace(typed))
		return lower == "1" || lower == "true" || lower == "yes" || lower == "on"
	case float64:
		return typed != 0
	case int:
		return typed != 0
	default:
		return false
	}
}

func asInt(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func asInt64(value interface{}) int64 {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func firstNonEmptyString(values ...string) string {
	for i := range values {
		if text := strings.TrimSpace(values[i]); text != "" {
			return text
		}
	}
	return ""
}
