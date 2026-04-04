package openclaw

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/cloud-claw-manager/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const (
	OpenClawInstanceLabelType        = "efucloud.com/claw.type"
	OpenClawInstanceLabelOwnerID     = "efucloud.com/claw.owner.id"
	OpenClawInstanceLabelOwner       = "efucloud.com/claw.owner"
	OpenClawInstanceLabelInstance    = "efucloud.com/claw.instance"
	OpenClawInstanceLabelTemplateRef = "efucloud.com/claw.template-ref"
	OpenClawInstanceLabelManagedBy   = "efucloud.com/claw.managed-by"
	openClawInstanceTypeValue        = "openclaw"
	openClawDefaultManagedByValue    = "cloud-claw-manager"
	openClawRestartedAtAnnotation    = "efucloud.com/claw.restartedAt"
	openClawInstanceStateAnnotation  = "efucloud.com/claw.instance.state"
	openClawGeneratedNameLength      = 6
	openClawGeneratedNamePrefix      = "openclaw-"
	openClawGatewayTokenLength       = 20
	openClawGatewayTokenSecretName   = "efucloud.com/claw.gateway-token.secret-name"
	openClawGatewayTokenSecretKey    = "efucloud.com/claw.gateway-token.secret-key"
	openClawGatewayTokenDataKey      = "OPENCLAW_GATEWAY_TOKEN"
	openClawEndpointAnnotation       = "efucloud.com/claw.endpoint"
	openClawDisplayNameAnnotation    = "efucloud.com/claw.display-name"
	openClawPurposeAnnotation        = "efucloud.com/claw.purpose"
	openClawOwnerAnnotation          = "efucloud.com/claw.owner"
	openClawOwnerIDAnnotation        = "efucloud.com/claw.owner.id"
	openClawOwnerUsernameAnnotation  = "efucloud.com/claw.owner.username"
	openClawOwnerEmailAnnotation     = "efucloud.com/claw.owner.email"
	openClawInstanceAnnotation       = "efucloud.com/claw.instance"
	openClawIngressHideHeadersAnno   = "nginx.ingress.kubernetes.io/proxy-hide-headers"
)

var (
	ErrOpenClawForbidden    = errors.New("forbidden")
	ErrOpenClawNotFound     = errors.New("not found")
	ErrOpenClawInvalidInput = errors.New("invalid input")
)

func legacyClawMetadataKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	return strings.Replace(key, "/claw.", "/openclaw.", 1)
}

func readMapValueWithLegacy(values map[string]string, key string) string {
	if len(values) == 0 {
		return ""
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if value := strings.TrimSpace(values[key]); value != "" {
		return value
	}
	legacyKey := legacyClawMetadataKey(key)
	if legacyKey != key {
		if value := strings.TrimSpace(values[legacyKey]); value != "" {
			return value
		}
	}
	return ""
}

func buildLegacyAwareLabelSelectors(key, value string) []string {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" {
		return nil
	}
	candidates := []string{
		key,
		legacyClawMetadataKey(key),
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(candidates))
	for _, candidateKey := range candidates {
		candidateKey = strings.TrimSpace(candidateKey)
		if candidateKey == "" {
			continue
		}
		selector := candidateKey
		if value != "" {
			selector = fmt.Sprintf("%s=%s", candidateKey, value)
		}
		if _, exists := seen[selector]; exists {
			continue
		}
		seen[selector] = struct{}{}
		out = append(out, selector)
	}
	return out
}

type OpenClawSecretRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type OpenClawProviderSpec struct {
	Name            string             `json:"name"`
	BaseURL         string             `json:"baseUrl,omitempty"`
	APIType         string             `json:"apiType,omitempty"`
	ModelIDs        []string           `json:"modelIds,omitempty"`
	APIKeySecretRef *OpenClawSecretRef `json:"apiKeySecretRef,omitempty"`
	Extra           map[string]string  `json:"extra,omitempty"`
}

type OpenClawModelsSpec struct {
	Primary   string                 `json:"primary,omitempty"`
	Providers []OpenClawProviderSpec `json:"providers,omitempty"`
}

type OpenClawAgentDefaults struct {
	ModelPrimary string `json:"modelPrimary,omitempty"`
}

type OpenClawAgentSpec struct {
	ID             string   `json:"id"`
	Name           string   `json:"name,omitempty"`
	Role           string   `json:"role,omitempty"`
	ModelPrimary   string   `json:"modelPrimary,omitempty"`
	Skills         []string `json:"skills,omitempty"`
	Channels       []string `json:"channels,omitempty"`
	SandboxProfile string   `json:"sandboxProfile,omitempty"`
	Enabled        bool     `json:"enabled"`
}

type OpenClawAgentsSpec struct {
	Defaults OpenClawAgentDefaults `json:"defaults,omitempty"`
	List     []OpenClawAgentSpec   `json:"list,omitempty"`
}

type OpenClawSkillRef struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type OpenClawSkillsSpec struct {
	Refs          []OpenClawSkillRef `json:"refs,omitempty"`
	InstallPolicy string             `json:"installPolicy,omitempty"`
}

type OpenClawGatewaySpec struct {
	AuthMode       string             `json:"authMode,omitempty"`
	TokenSecretRef *OpenClawSecretRef `json:"tokenSecretRef,omitempty"`
}

type OpenClawRuntimeSpec struct {
	Image     string `json:"image,omitempty"`
	Replicas  *int32 `json:"replicas,omitempty"`
	Paused    bool   `json:"paused,omitempty"`
	RestartAt string `json:"restartAt,omitempty"`
}

type OpenClawTemplateSpec struct {
	Resources         []runtime.RawExtension `json:"resources,omitempty"`
	ConfigMapSelector string                 `json:"configMapSelector,omitempty"`
}

type OpenClawAppliedSkillStatus struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Digest  string `json:"digest,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

type OpenClawInstanceStatus struct {
	Phase              string                       `json:"phase,omitempty"`
	Message            string                       `json:"message,omitempty"`
	Endpoint           string                       `json:"endpoint,omitempty"`
	GatewayToken       string                       `json:"gatewayToken,omitempty"`
	ReadyReplicas      int32                        `json:"readyReplicas,omitempty"`
	ObservedGeneration int64                        `json:"observedGeneration,omitempty"`
	LastTransitionAt   string                       `json:"lastTransitionAt,omitempty"`
	AppliedSkills      []OpenClawAppliedSkillStatus `json:"appliedSkills,omitempty"`
}

type OpenClawInstanceCreateRequest struct {
	Namespace       string                 `json:"namespace"`
	Name            string                 `json:"name"`
	DisplayName     string                 `json:"displayName,omitempty"`
	Purpose         string                 `json:"purpose,omitempty"`
	VisibilityUsers []string               `json:"visibilityUsers,omitempty"`
	TemplateRef     string                 `json:"templateRef,omitempty"`
	TemplateValues  map[string]interface{} `json:"templateValues,omitempty"`
	Runtime         OpenClawRuntimeSpec    `json:"runtime,omitempty"`
	Template        OpenClawTemplateSpec   `json:"template,omitempty"`
	Gateway         OpenClawGatewaySpec    `json:"gateway,omitempty"`
	Models          OpenClawModelsSpec     `json:"models,omitempty"`
	Agents          OpenClawAgentsSpec     `json:"agents,omitempty"`
	Skills          OpenClawSkillsSpec     `json:"skills,omitempty"`
}

type OpenClawInstance struct {
	Namespace       string                 `json:"namespace"`
	Name            string                 `json:"name"`
	DisplayName     string                 `json:"displayName,omitempty"`
	Purpose         string                 `json:"purpose,omitempty"`
	OwnerID         string                 `json:"ownerId"`
	OwnerUsername   string                 `json:"ownerUsername,omitempty"`
	OwnerEmail      string                 `json:"ownerEmail,omitempty"`
	VisibilityUsers []string               `json:"visibilityUsers,omitempty"`
	Accessible      bool                   `json:"accessible"`
	Runtime         OpenClawRuntimeSpec    `json:"runtime,omitempty"`
	Template        OpenClawTemplateSpec   `json:"template,omitempty"`
	Gateway         OpenClawGatewaySpec    `json:"gateway,omitempty"`
	Models          OpenClawModelsSpec     `json:"models,omitempty"`
	Agents          OpenClawAgentsSpec     `json:"agents,omitempty"`
	Skills          OpenClawSkillsSpec     `json:"skills,omitempty"`
	Status          OpenClawInstanceStatus `json:"status,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	Annotations     map[string]string      `json:"annotations,omitempty"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

type OpenClawInstanceListResponse struct {
	Total uint               `json:"total"`
	Data  []OpenClawInstance `json:"data"`
}

type OpenClawInstanceActionResponse struct {
	Action  string           `json:"action"`
	Message string           `json:"message"`
	Data    OpenClawInstance `json:"data"`
}

type OpenClawTemplate struct {
	Name        string                 `json:"name"`
	Namespace   string                 `json:"namespace"`
	Description string                 `json:"description,omitempty"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Defaults    map[string]interface{} `json:"defaults,omitempty"`
	Templates   string                 `json:"templates,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

type OpenClawTemplateListResponse struct {
	Total uint               `json:"total"`
	Data  []OpenClawTemplate `json:"data"`
}

type OpenClawTemplateUpsertRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Schema      map[string]interface{} `json:"schema"`
	Defaults    map[string]interface{} `json:"defaults,omitempty"`
	Templates   string                 `json:"templates"`
}

type OpenClawDashboardResponse struct {
	Overview  OpenClawOverview            `json:"overview"`
	Providers []OpenClawProviderBreakdown `json:"providers"`
	Instances []OpenClawInstanceTelemetry `json:"instances"`
}

type OpenClawOverview struct {
	TotalInstances      int     `json:"totalInstances"`
	AccessibleInstances int     `json:"accessibleInstances"`
	ReadyPods           int     `json:"readyPods"`
	MemoryBytes         float64 `json:"memoryBytes"`
	NetworkRxBytesPS    float64 `json:"networkRxBytesPerSecond"`
	NetworkTxBytesPS    float64 `json:"networkTxBytesPerSecond"`
	InputTokens24h      float64 `json:"inputTokens24h"`
	OutputTokens24h     float64 `json:"outputTokens24h"`
	CostUSD24h          float64 `json:"costUSD24h"`
}

type OpenClawProviderBreakdown struct {
	Provider string  `json:"provider"`
	Requests float64 `json:"requests"`
}

type OpenClawInstanceTelemetry struct {
	Namespace          string                   `json:"namespace"`
	Name               string                   `json:"name"`
	DisplayName        string                   `json:"displayName,omitempty"`
	Purpose            string                   `json:"purpose,omitempty"`
	OwnerID            string                   `json:"ownerId,omitempty"`
	OwnerUsername      string                   `json:"ownerUsername,omitempty"`
	OwnerEmail         string                   `json:"ownerEmail,omitempty"`
	Accessible         bool                     `json:"accessible"`
	ReadyPods          int                      `json:"readyPods"`
	Provider           string                   `json:"provider,omitempty"`
	Endpoint           string                   `json:"endpoint,omitempty"`
	GatewayToken       string                   `json:"gatewayToken,omitempty"`
	ProviderCandidates []string                 `json:"providerCandidates,omitempty"`
	ProviderPrimary    string                   `json:"providerPrimary,omitempty"`
	ProviderModelIDs   []string                 `json:"providerModelIds,omitempty"`
	ProviderBaseURL    string                   `json:"providerBaseUrl,omitempty"`
	ProviderAPIType    string                   `json:"providerApiType,omitempty"`
	MemoryBytes        float64                  `json:"memoryBytes"`
	NetworkRxBytesPS   float64                  `json:"networkRxBytesPerSecond"`
	NetworkTxBytesPS   float64                  `json:"networkTxBytesPerSecond"`
	InputTokens24h     float64                  `json:"inputTokens24h"`
	OutputTokens24h    float64                  `json:"outputTokens24h"`
	CostUSD24h         float64                  `json:"costUSD24h"`
	RuntimeInsights    *OpenClawRuntimeInsights `json:"runtimeInsights,omitempty"`
	UpdatedAt          string                   `json:"updatedAt,omitempty"`
}

type OpenClawRuntimeInsights struct {
	CollectedAt                string   `json:"collectedAt,omitempty"`
	SourcePod                  string   `json:"sourcePod,omitempty"`
	SourceContainer            string   `json:"sourceContainer,omitempty"`
	StatusState                string   `json:"statusState,omitempty"`
	GatewayState               string   `json:"gatewayState,omitempty"`
	SecurityState              string   `json:"securityState,omitempty"`
	RuntimeVersion             string   `json:"runtimeVersion,omitempty"`
	OpenClawConfigPath         string   `json:"openclawConfigPath,omitempty"`
	OpenClawAgentsCount        int      `json:"openclawAgentsCount,omitempty"`
	OpenClawProvidersCount     int      `json:"openclawProvidersCount,omitempty"`
	OpenClawPrimaryModel       string   `json:"openclawPrimaryModel,omitempty"`
	OpenClawWorkspace          string   `json:"openclawWorkspace,omitempty"`
	ControlUiDisableDeviceAuth *bool    `json:"controlUiDisableDeviceAuth,omitempty"`
	ControlUiAllowedOrigins    []string `json:"controlUiAllowedOrigins,omitempty"`
	CronJobsCount              int      `json:"cronJobsCount,omitempty"`
	SessionIndexFilesCount     int      `json:"sessionIndexFilesCount,omitempty"`
	SessionCount               int      `json:"sessionCount,omitempty"`
	SessionLastUpdatedAt       string   `json:"sessionLastUpdatedAt,omitempty"`
	LastError                  string   `json:"lastError,omitempty"`
}

type openClawInstanceState struct {
	VisibilityUsers []string             `json:"visibilityUsers,omitempty"`
	Runtime         OpenClawRuntimeSpec  `json:"runtime,omitempty"`
	Gateway         OpenClawGatewaySpec  `json:"gateway,omitempty"`
	Models          OpenClawModelsSpec   `json:"models,omitempty"`
	Agents          OpenClawAgentsSpec   `json:"agents,omitempty"`
	Skills          OpenClawSkillsSpec   `json:"skills,omitempty"`
	Template        OpenClawTemplateSpec `json:"template,omitempty"`
}

type OpenClawOwnerMeta struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}

type openClawClients struct {
	mu        sync.Mutex
	inited    bool
	kube      *kubernetes.Clientset
	dynamic   dynamic.Interface
	discovery *discovery.DiscoveryClient
	err       error
}

var defaultOpenClawClients = &openClawClients{}

type OpenClawInstanceService struct{}

func (s OpenClawInstanceService) ListInstances(ctx context.Context, requesterID, namespace string) (OpenClawInstanceListResponse, error) {
	kc, _, _, err := defaultOpenClawClients.get(ctx)
	if err != nil {
		return OpenClawInstanceListResponse{}, err
	}
	items, err := listInstanceDeployments(ctx, kc, strings.TrimSpace(namespace), "")
	if err != nil {
		return OpenClawInstanceListResponse{}, err
	}
	instances := aggregateInstances(items, requesterID)
	enrichInstanceModelProviders(ctx, kc, instances)
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].CreatedAt.After(instances[j].CreatedAt)
	})
	enrichInstanceGatewayTokens(ctx, kc, instances, requesterID)
	return OpenClawInstanceListResponse{Total: uint(len(instances)), Data: instances}, nil
}

func (s OpenClawInstanceService) GetInstance(ctx context.Context, requesterID, namespace, name string) (OpenClawInstance, error) {
	kc, _, _, err := defaultOpenClawClients.get(ctx)
	if err != nil {
		return OpenClawInstance{}, err
	}
	items, err := listInstanceDeployments(ctx, kc, strings.TrimSpace(namespace), strings.TrimSpace(name))
	if err != nil {
		return OpenClawInstance{}, err
	}
	instances := aggregateInstances(items, requesterID)
	if len(instances) == 0 {
		return OpenClawInstance{}, ErrOpenClawNotFound
	}
	instance := instances[0]
	enrichInstanceModelProviders(ctx, kc, instances)
	enrichInstanceGatewayTokens(ctx, kc, instances, requesterID)
	if !instance.Accessible {
		return OpenClawInstance{}, ErrOpenClawForbidden
	}
	return instances[0], nil
}

func (s OpenClawInstanceService) CreateInstance(ctx context.Context, requesterID string, ownerMeta OpenClawOwnerMeta, req OpenClawInstanceCreateRequest) (OpenClawInstance, error) {
	if strings.TrimSpace(requesterID) == "" {
		return OpenClawInstance{}, fmt.Errorf("%w: requester is required", ErrOpenClawInvalidInput)
	}
	ownerMeta.Username = strings.TrimSpace(ownerMeta.Username)
	ownerMeta.Email = strings.TrimSpace(ownerMeta.Email)
	namespace := strings.TrimSpace(req.Namespace)
	if namespace == "" {
		return OpenClawInstance{}, fmt.Errorf("%w: namespace is required", ErrOpenClawInvalidInput)
	}
	if errs := validation.IsDNS1123Label(namespace); len(errs) > 0 {
		return OpenClawInstance{}, fmt.Errorf("%w: invalid namespace: %s", ErrOpenClawInvalidInput, strings.Join(errs, ","))
	}

	kc, dc, disco, err := defaultOpenClawClients.get(ctx)
	if err != nil {
		return OpenClawInstance{}, err
	}
	if _, err = kc.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return OpenClawInstance{}, fmt.Errorf("%w: namespace %s not found", ErrOpenClawInvalidInput, namespace)
		}
		return OpenClawInstance{}, err
	}
	if strings.TrimSpace(req.Name) != "" {
		config.Logger.Infof("create instance request provided name %q, ignored because server generates random instance name", strings.TrimSpace(req.Name))
	}
	name, err := generateInstanceName(ctx, kc, namespace)
	if err != nil {
		return OpenClawInstance{}, err
	}

	runtimeSpec := req.Runtime
	// OpenClaw 实例副本数固定为 1，不允许通过配置或请求覆盖。
	r := int32(1)
	runtimeSpec.Replicas = &r
	runtimeSpec.Image = strings.TrimSpace(runtimeSpec.Image)
	if runtimeSpec.Paused {
		r := int32(0)
		runtimeSpec.Replicas = &r
	}

	templateSpec, err := normalizeTemplate(req.Template)
	if err != nil {
		return OpenClawInstance{}, err
	}
	templateValues := map[string]interface{}{}
	for k, v := range req.TemplateValues {
		templateValues[k] = v
	}
	injectPlatformTemplateValues(templateValues)
	if _, ok := templateValues["instanceName"]; !ok {
		templateValues["instanceName"] = name
	}
	if _, ok := templateValues["name"]; !ok {
		templateValues["name"] = name
	}
	if _, ok := templateValues["namespace"]; !ok {
		templateValues["namespace"] = namespace
	}
	if runtimeSpec.Image == "" {
		if image := strings.TrimSpace(fmt.Sprintf("%v", templateValues["image"])); image != "" {
			runtimeSpec.Image = image
		}
	}
	displayName := resolveTextWithAliases(strings.TrimSpace(req.DisplayName), templateValues, []string{"displayName", "openclawName", "openclawDisplayName"})
	purpose := resolveTextWithAliases(strings.TrimSpace(req.Purpose), templateValues, []string{"purpose", "openclawPurpose", "description"})
	if displayName != "" {
		templateValues["displayName"] = displayName
		templateValues["openclawName"] = displayName
	}
	if purpose != "" {
		templateValues["purpose"] = purpose
		templateValues["openclawPurpose"] = purpose
	}
	gatewayToken, err := generateGatewayToken()
	if err != nil {
		return OpenClawInstance{}, err
	}
	templateValues["gatewayToken"] = gatewayToken
	templateValues["openclawGatewayToken"] = gatewayToken
	templateValues["OPENCLAW_GATEWAY_TOKEN"] = gatewayToken
	// ownerId 由认证上下文注入，不能依赖用户输入
	templateValues["ownerId"] = requesterID
	templateValues["ownerID"] = requesterID
	templateValues["ownerUsername"] = ownerMeta.Username
	templateValues["ownerEmail"] = ownerMeta.Email
	previewEndpoint := buildPreviewEndpoint(namespace, name, config.ApplicationConfig.OpenClawControl.PreviewBaseDomain)
	templateValues["previewEndpoint"] = previewEndpoint
	// Ingress 模版可直接引用 host，避免在模板中做 URL 解析。
	templateValues["previewHost"] = extractEndpointHost(previewEndpoint)
	templateRef := strings.TrimSpace(req.TemplateRef)
	if templateRef != "" {
		rendered, renderErr := (OpenClawTemplateService{}).RenderTemplate(ctx, templateRef, templateValues)
		if renderErr != nil {
			return OpenClawInstance{}, renderErr
		}
		templateSpec.Resources = append(templateSpec.Resources, rendered...)
	}
	if templateSpec.ConfigMapSelector != "" {
		resources, loadErr := loadTemplateResourcesFromConfigMaps(ctx, kc, namespace, templateSpec.ConfigMapSelector)
		if loadErr != nil {
			return OpenClawInstance{}, loadErr
		}
		templateSpec.Resources = append(templateSpec.Resources, resources...)
	}
	if len(templateSpec.Resources) == 0 {
		return OpenClawInstance{}, fmt.Errorf("%w: template resources are required", ErrOpenClawInvalidInput)
	}
	if previewEndpoint := strings.TrimSpace(fmt.Sprintf("%v", templateValues["previewEndpoint"])); previewEndpoint != "" {
		templateSpec.Resources, err = injectDeploymentEndpointAnnotations(templateSpec.Resources, previewEndpoint)
		if err != nil {
			return OpenClawInstance{}, err
		}
	}
	templateSpec.Resources, err = injectIngressPreviewAnnotations(templateSpec.Resources)
	if err != nil {
		return OpenClawInstance{}, err
	}

	gatewayTokenRef := normalizeSecretRef(req.Gateway.TokenSecretRef)
	if gatewayTokenRef == nil {
		gatewayTokenRef = inferGatewayTokenSecretRefFromTemplateResources(templateSpec.Resources)
	}
	if gatewayTokenRef != nil {
		req.Gateway.TokenSecretRef = gatewayTokenRef
	}
	templateSpec.Resources, err = InjectOpenClawOwnershipLabels(templateSpec.Resources, requesterID, name, templateRef)
	if err != nil {
		return OpenClawInstance{}, err
	}

	state := openClawInstanceState{
		VisibilityUsers: normalizeUsers(req.VisibilityUsers),
		Runtime:         runtimeSpec,
		Template:        OpenClawTemplateSpec{ConfigMapSelector: templateSpec.ConfigMapSelector},
		Gateway:         req.Gateway,
		Models:          normalizeModels(req.Models),
		Agents:          normalizeAgents(req.Agents),
		Skills:          normalizeSkills(req.Skills),
	}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return OpenClawInstance{}, err
	}
	resourceAnnotations := map[string]string{
		openClawInstanceStateAnnotation: string(stateJSON),
		// 统一注入 owner/instance 注解，便于通过 annotations 进行审计与排障。
		openClawOwnerAnnotation:    requesterID,
		openClawOwnerIDAnnotation:  requesterID,
		openClawInstanceAnnotation: name,
	}
	if ownerMeta.Username != "" {
		resourceAnnotations[openClawOwnerUsernameAnnotation] = ownerMeta.Username
	}
	if ownerMeta.Email != "" {
		resourceAnnotations[openClawOwnerEmailAnnotation] = ownerMeta.Email
	}
	if displayName != "" {
		resourceAnnotations[openClawDisplayNameAnnotation] = displayName
	}
	if purpose != "" {
		resourceAnnotations[openClawPurposeAnnotation] = purpose
	}
	if gatewayTokenRef != nil {
		resourceAnnotations[openClawGatewayTokenSecretName] = gatewayTokenRef.Name
		resourceAnnotations[openClawGatewayTokenSecretKey] = gatewayTokenRef.Key
	}
	resourceLabels := map[string]string{
		OpenClawInstanceLabelType:      openClawInstanceTypeValue,
		OpenClawInstanceLabelOwnerID:   requesterID,
		OpenClawInstanceLabelOwner:     requesterID,
		OpenClawInstanceLabelInstance:  name,
		OpenClawInstanceLabelManagedBy: openClawDefaultManagedByValue,
	}
	if templateRef != "" {
		resourceLabels[OpenClawInstanceLabelTemplateRef] = templateRef
	}
	if err = applyTemplateResources(ctx, dc, disco, namespace, templateSpec.Resources, resourceLabels, resourceAnnotations); err != nil {
		return OpenClawInstance{}, err
	}
	return s.GetInstance(ctx, requesterID, namespace, name)
}

func (s OpenClawInstanceService) DeleteInstance(ctx context.Context, requesterID, namespace, name string) error {
	instance, err := s.GetInstance(ctx, requesterID, namespace, name)
	if err != nil {
		return err
	}
	_, dc, disco, err := defaultOpenClawClients.get(ctx)
	if err != nil {
		return err
	}
	return deleteInstanceResources(ctx, dc, disco, instance.Namespace, instance.Name, instance.OwnerID)
}

func (s OpenClawInstanceService) ControlInstance(ctx context.Context, requesterID, namespace, name, action string) (OpenClawInstance, error) {
	instance, err := s.GetInstance(ctx, requesterID, namespace, name)
	if err != nil {
		return OpenClawInstance{}, err
	}
	kc, _, _, err := defaultOpenClawClients.get(ctx)
	if err != nil {
		return OpenClawInstance{}, err
	}
	items, err := listInstanceDeployments(ctx, kc, instance.Namespace, instance.Name)
	if err != nil {
		return OpenClawInstance{}, err
	}
	if len(items) == 0 {
		return OpenClawInstance{}, ErrOpenClawNotFound
	}

	act := strings.ToLower(strings.TrimSpace(action))
	now := time.Now().UTC().Format(time.RFC3339)
	targetReplicas := int32(1)

	for i := range items {
		dep := items[i].DeepCopy()
		switch act {
		case "start":
			dep.Spec.Replicas = int32Ptr(targetReplicas)
			updateDeploymentState(dep, func(state *openClawInstanceState) {
				state.Runtime.Paused = false
				state.Runtime.Replicas = int32Ptr(targetReplicas)
			})
		case "stop":
			dep.Spec.Replicas = int32Ptr(0)
			updateDeploymentState(dep, func(state *openClawInstanceState) {
				state.Runtime.Paused = true
				state.Runtime.Replicas = int32Ptr(0)
			})
		case "restart":
			if dep.Spec.Template.Annotations == nil {
				dep.Spec.Template.Annotations = map[string]string{}
			}
			dep.Spec.Template.Annotations[openClawRestartedAtAnnotation] = now
			updateDeploymentState(dep, func(state *openClawInstanceState) {
				state.Runtime.RestartAt = now
			})
		default:
			return OpenClawInstance{}, fmt.Errorf("%w: unsupported action %q", ErrOpenClawInvalidInput, action)
		}
		if _, err = kc.AppsV1().Deployments(dep.Namespace).Update(ctx, dep, metav1.UpdateOptions{}); err != nil {
			return OpenClawInstance{}, err
		}
	}
	return s.GetInstance(ctx, requesterID, instance.Namespace, instance.Name)
}

func (c *openClawClients) get(ctx context.Context) (*kubernetes.Clientset, dynamic.Interface, *discovery.DiscoveryClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.inited {
		return c.kube, c.dynamic, c.discovery, c.err
	}
	restCfg, err := kube.BuildRestConfig(ctx, config.RunKubeConfig)
	if err != nil {
		c.err = err
		c.inited = true
		return nil, nil, nil, err
	}
	kc, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		c.err = err
		c.inited = true
		return nil, nil, nil, err
	}
	dc, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		c.err = err
		c.inited = true
		return nil, nil, nil, err
	}
	disco, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		c.err = err
		c.inited = true
		return nil, nil, nil, err
	}
	c.kube = kc
	c.dynamic = dc
	c.discovery = disco
	c.inited = true
	return c.kube, c.dynamic, c.discovery, nil
}

func listInstanceDeployments(ctx context.Context, kc *kubernetes.Clientset, namespace, instanceName string) ([]appsv1.Deployment, error) {
	selectors := buildLegacyAwareLabelSelectors(OpenClawInstanceLabelInstance, instanceName)
	out := make([]appsv1.Deployment, 0)
	seen := map[string]struct{}{}
	for _, selector := range selectors {
		list, err := kc.AppsV1().Deployments(strings.TrimSpace(namespace)).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return nil, err
		}
		for i := range list.Items {
			key := list.Items[i].Namespace + "/" + list.Items[i].Name
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, list.Items[i])
		}
	}
	return out, nil
}

func aggregateInstances(deployments []appsv1.Deployment, requesterID string) []OpenClawInstance {
	groups := map[string][]appsv1.Deployment{}
	for i := range deployments {
		dep := deployments[i]
		instanceName := readMapValueWithLegacy(dep.Labels, OpenClawInstanceLabelInstance)
		if instanceName == "" {
			instanceName = dep.Name
		}
		key := dep.Namespace + "/" + instanceName
		groups[key] = append(groups[key], dep)
	}
	out := make([]OpenClawInstance, 0, len(groups))
	for _, group := range groups {
		item, ok := buildInstanceFromDeployments(group, requesterID)
		if !ok || !item.Accessible {
			continue
		}
		out = append(out, item)
	}
	return out
}

func buildInstanceFromDeployments(group []appsv1.Deployment, requesterID string) (OpenClawInstance, bool) {
	if len(group) == 0 {
		return OpenClawInstance{}, false
	}
	sort.Slice(group, func(i, j int) bool {
		return group[i].CreationTimestamp.Time.Before(group[j].CreationTimestamp.Time)
	})
	base := group[0]
	instanceName := readMapValueWithLegacy(base.Labels, OpenClawInstanceLabelInstance)
	if instanceName == "" {
		instanceName = readMapValueWithLegacy(base.Annotations, openClawInstanceAnnotation)
	}
	if instanceName == "" {
		instanceName = base.Name
	}
	owner := readMapValueWithLegacy(base.Labels, OpenClawInstanceLabelOwner)
	if owner == "" {
		owner = readMapValueWithLegacy(base.Labels, OpenClawInstanceLabelOwnerID)
	}
	if owner == "" {
		owner = readMapValueWithLegacy(base.Annotations, openClawOwnerAnnotation)
	}
	if owner == "" {
		owner = readMapValueWithLegacy(base.Annotations, openClawOwnerIDAnnotation)
	}
	ownerUsername := detectInstanceAnnotation(group, openClawOwnerUsernameAnnotation)
	ownerEmail := detectInstanceAnnotation(group, openClawOwnerEmailAnnotation)
	state := parseInstanceState(base.Annotations)
	readyReplicas := int32(0)
	createdAt := base.CreationTimestamp.Time
	updatedAt := deploymentUpdatedAt(base)
	for i := range group {
		readyReplicas += group[i].Status.ReadyReplicas
		if group[i].CreationTimestamp.Time.Before(createdAt) {
			createdAt = group[i].CreationTimestamp.Time
		}
		itemUpdated := deploymentUpdatedAt(group[i])
		if itemUpdated.After(updatedAt) {
			updatedAt = itemUpdated
		}
	}

	runtimeSpec := state.Runtime
	if runtimeSpec.Replicas == nil {
		if base.Spec.Replicas != nil {
			r := *base.Spec.Replicas
			runtimeSpec.Replicas = &r
		}
	}
	if runtimeSpec.Image == "" && len(base.Spec.Template.Spec.Containers) > 0 {
		runtimeSpec.Image = strings.TrimSpace(base.Spec.Template.Spec.Containers[0].Image)
	}
	if runtimeSpec.Replicas != nil {
		runtimeSpec.Paused = *runtimeSpec.Replicas == 0
	}

	accessible := isAccessible(owner, state.VisibilityUsers, requesterID)
	displayName := detectInstanceAnnotation(group, openClawDisplayNameAnnotation)
	purpose := detectInstanceAnnotation(group, openClawPurposeAnnotation)
	status := OpenClawInstanceStatus{
		ReadyReplicas: readyReplicas,
		Endpoint:      detectInstanceEndpoint(group),
	}
	if readyReplicas > 0 {
		status.Phase = "Running"
	} else {
		status.Phase = "Stopped"
	}
	return OpenClawInstance{
		Namespace:       base.Namespace,
		Name:            instanceName,
		DisplayName:     displayName,
		Purpose:         purpose,
		OwnerID:         owner,
		OwnerUsername:   ownerUsername,
		OwnerEmail:      ownerEmail,
		VisibilityUsers: state.VisibilityUsers,
		Accessible:      accessible,
		Runtime:         runtimeSpec,
		Template:        state.Template,
		Gateway:         state.Gateway,
		Models:          state.Models,
		Agents:          state.Agents,
		Skills:          state.Skills,
		Status:          status,
		Labels:          base.Labels,
		Annotations:     base.Annotations,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, true
}

func deploymentUpdatedAt(dep appsv1.Deployment) time.Time {
	updated := dep.CreationTimestamp.Time
	for i := range dep.Status.Conditions {
		cond := dep.Status.Conditions[i]
		if cond.LastUpdateTime.After(updated) {
			updated = cond.LastUpdateTime.Time
		}
		if cond.LastTransitionTime.After(updated) {
			updated = cond.LastTransitionTime.Time
		}
	}
	return updated
}

func detectInstanceEndpoint(group []appsv1.Deployment) string {
	keys := []string{
		openClawEndpointAnnotation,
		legacyClawMetadataKey(openClawEndpointAnnotation),
		"openclaw.efucloud.com/endpoint",
		"openclaw.endpoint",
		"endpoint",
	}
	for i := range group {
		for _, key := range keys {
			if value := readMapValueWithLegacy(group[i].Annotations, key); value != "" {
				return value
			}
			if value := readMapValueWithLegacy(group[i].Spec.Template.Annotations, key); value != "" {
				return value
			}
		}
	}
	return ""
}

func detectInstanceAnnotation(group []appsv1.Deployment, key string) string {
	annotationKey := strings.TrimSpace(key)
	if annotationKey == "" {
		return ""
	}
	for i := range group {
		if value := readMapValueWithLegacy(group[i].Annotations, annotationKey); value != "" {
			return value
		}
		if value := readMapValueWithLegacy(group[i].Spec.Template.Annotations, annotationKey); value != "" {
			return value
		}
	}
	return ""
}

func injectDeploymentEndpointAnnotations(resources []runtime.RawExtension, endpoint string) ([]runtime.RawExtension, error) {
	previewEndpoint := strings.TrimSpace(endpoint)
	if previewEndpoint == "" || len(resources) == 0 {
		return resources, nil
	}
	out := make([]runtime.RawExtension, 0, len(resources))
	for i := range resources {
		raw := bytes.TrimSpace(resources[i].Raw)
		if len(raw) == 0 {
			continue
		}
		obj := map[string]interface{}{}
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, fmt.Errorf("%w: template resource[%d] is not valid JSON: %v", ErrOpenClawInvalidInput, i, err)
		}
		kind := strings.TrimSpace(fmt.Sprintf("%v", obj["kind"]))
		if strings.EqualFold(kind, "Deployment") {
			meta := ensureObjectMap(obj, "metadata")
			annotations := ensureStringMap(meta, "annotations")
			annotations[openClawEndpointAnnotation] = previewEndpoint
			meta["annotations"] = annotations
			obj["metadata"] = meta

			spec := ensureObjectMap(obj, "spec")
			templateObj := ensureObjectMap(spec, "template")
			templateMeta := ensureObjectMap(templateObj, "metadata")
			templateAnnos := ensureStringMap(templateMeta, "annotations")
			templateAnnos[openClawEndpointAnnotation] = previewEndpoint
			templateMeta["annotations"] = templateAnnos
			templateObj["metadata"] = templateMeta
			spec["template"] = templateObj
			obj["spec"] = spec
		}
		nextRaw, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		out = append(out, runtime.RawExtension{Raw: nextRaw})
	}
	return out, nil
}

func injectIngressPreviewAnnotations(resources []runtime.RawExtension) ([]runtime.RawExtension, error) {
	if len(resources) == 0 {
		return resources, nil
	}
	out := make([]runtime.RawExtension, 0, len(resources))
	for i := range resources {
		raw := bytes.TrimSpace(resources[i].Raw)
		if len(raw) == 0 {
			continue
		}
		obj := map[string]interface{}{}
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, fmt.Errorf("%w: template resource[%d] is not valid JSON: %v", ErrOpenClawInvalidInput, i, err)
		}
		kind := strings.TrimSpace(fmt.Sprintf("%v", obj["kind"]))
		if strings.EqualFold(kind, "Ingress") {
			meta := ensureObjectMap(obj, "metadata")
			annotations := ensureStringMap(meta, "annotations")
			annotations[openClawIngressHideHeadersAnno] = mergeCSVAnnotationValue(
				annotations[openClawIngressHideHeadersAnno],
				[]string{"X-Frame-Options", "Content-Security-Policy"},
			)
			meta["annotations"] = annotations
			obj["metadata"] = meta
		}
		nextRaw, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		out = append(out, runtime.RawExtension{Raw: nextRaw})
	}
	return out, nil
}

func mergeCSVAnnotationValue(current string, required []string) string {
	values := make([]string, 0, len(required)+2)
	seen := map[string]struct{}{}
	appendValue := func(raw string) {
		item := strings.TrimSpace(raw)
		if item == "" {
			return
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		values = append(values, item)
	}
	for _, item := range strings.Split(current, ",") {
		appendValue(item)
	}
	for _, item := range required {
		appendValue(item)
	}
	return strings.Join(values, ",")
}

func parseInstanceState(annotations map[string]string) openClawInstanceState {
	if len(annotations) == 0 {
		return openClawInstanceState{}
	}
	raw := readMapValueWithLegacy(annotations, openClawInstanceStateAnnotation)
	if raw == "" {
		return openClawInstanceState{}
	}
	var state openClawInstanceState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return openClawInstanceState{}
	}
	state.VisibilityUsers = normalizeUsers(state.VisibilityUsers)
	state.Models = normalizeModels(state.Models)
	state.Agents = normalizeAgents(state.Agents)
	state.Skills = normalizeSkills(state.Skills)
	state.Template.Resources = nil
	return state
}

func updateDeploymentState(dep *appsv1.Deployment, mutate func(state *openClawInstanceState)) {
	state := parseInstanceState(dep.Annotations)
	mutate(&state)
	payload, err := json.Marshal(state)
	if err != nil {
		return
	}
	if dep.Annotations == nil {
		dep.Annotations = map[string]string{}
	}
	dep.Annotations[openClawInstanceStateAnnotation] = string(payload)
}

func isAccessible(owner string, visibilityUsers []string, requesterID string) bool {
	requester := strings.TrimSpace(requesterID)
	if requester == "" {
		return false
	}
	if strings.TrimSpace(owner) == requester {
		return true
	}
	for _, user := range visibilityUsers {
		if strings.TrimSpace(user) == requester {
			return true
		}
	}
	return false
}

func applyTemplateResources(
	ctx context.Context,
	dc dynamic.Interface,
	disco *discovery.DiscoveryClient,
	namespace string,
	resources []runtime.RawExtension,
	labels map[string]string,
	annotations map[string]string,
) error {
	// Owner/Instance 元信息需要覆盖到所有资源与 workload PodTemplate，保证后续筛选一致。
	identityLabels := map[string]string{
		OpenClawInstanceLabelType:     strings.TrimSpace(labels[OpenClawInstanceLabelType]),
		OpenClawInstanceLabelOwner:    strings.TrimSpace(labels[OpenClawInstanceLabelOwner]),
		OpenClawInstanceLabelOwnerID:  strings.TrimSpace(labels[OpenClawInstanceLabelOwnerID]),
		OpenClawInstanceLabelInstance: strings.TrimSpace(labels[OpenClawInstanceLabelInstance]),
	}
	if templateRef := strings.TrimSpace(labels[OpenClawInstanceLabelTemplateRef]); templateRef != "" {
		identityLabels[OpenClawInstanceLabelTemplateRef] = templateRef
	}
	identityAnnotations := map[string]string{
		openClawOwnerAnnotation:         strings.TrimSpace(annotations[openClawOwnerAnnotation]),
		openClawOwnerIDAnnotation:       strings.TrimSpace(annotations[openClawOwnerIDAnnotation]),
		openClawOwnerUsernameAnnotation: strings.TrimSpace(annotations[openClawOwnerUsernameAnnotation]),
		openClawOwnerEmailAnnotation:    strings.TrimSpace(annotations[openClawOwnerEmailAnnotation]),
		openClawInstanceAnnotation:      strings.TrimSpace(annotations[openClawInstanceAnnotation]),
	}
	if identityAnnotations[openClawOwnerAnnotation] == "" {
		identityAnnotations[openClawOwnerAnnotation] = identityLabels[OpenClawInstanceLabelOwner]
	}
	if identityAnnotations[openClawOwnerIDAnnotation] == "" {
		identityAnnotations[openClawOwnerIDAnnotation] = identityLabels[OpenClawInstanceLabelOwnerID]
	}
	if identityAnnotations[openClawInstanceAnnotation] == "" {
		identityAnnotations[openClawInstanceAnnotation] = identityLabels[OpenClawInstanceLabelInstance]
	}

	for i := range resources {
		raw := bytes.TrimSpace(resources[i].Raw)
		if len(raw) == 0 {
			continue
		}
		obj := map[string]interface{}{}
		if err := json.Unmarshal(raw, &obj); err != nil {
			return fmt.Errorf("%w: invalid resource json at index %d: %v", ErrOpenClawInvalidInput, i, err)
		}
		uObj := &unstructured.Unstructured{Object: obj}
		apiVersion := strings.TrimSpace(uObj.GetAPIVersion())
		kind := strings.TrimSpace(uObj.GetKind())
		if apiVersion == "" || kind == "" {
			return fmt.Errorf("%w: template resource[%d] missing apiVersion/kind", ErrOpenClawInvalidInput, i)
		}
		gv, err := schema.ParseGroupVersion(apiVersion)
		if err != nil {
			return err
		}
		gvr, namespaced, err := resolveTemplateResourceGVR(ctx, disco, gv, kind)
		if err != nil {
			return err
		}
		metaObj := ensureObjectMap(obj, "metadata")
		metaLabels := ensureStringMap(metaObj, "labels")
		for k, v := range labels {
			if strings.TrimSpace(v) != "" {
				metaLabels[k] = v
			}
		}
		metaObj["labels"] = metaLabels
		metaAnnos := ensureStringMap(metaObj, "annotations")
		for k, v := range annotations {
			if strings.TrimSpace(v) != "" {
				metaAnnos[k] = v
			}
		}
		metaObj["annotations"] = metaAnnos
		obj["metadata"] = metaObj
		injectWorkloadTemplateIdentity(obj, identityLabels, identityAnnotations)
		uObj = &unstructured.Unstructured{Object: obj}
		if strings.TrimSpace(uObj.GetName()) == "" {
			return fmt.Errorf("%w: template resource[%d] missing metadata.name", ErrOpenClawInvalidInput, i)
		}
		var res dynamic.ResourceInterface
		if namespaced {
			uObj.SetNamespace(namespace)
			res = dc.Resource(gvr).Namespace(namespace)
		} else {
			res = dc.Resource(gvr)
		}
		payload, err := json.Marshal(uObj.Object)
		if err != nil {
			return err
		}
		force := true
		if _, err = res.Patch(
			ctx,
			uObj.GetName(),
			types.ApplyPatchType,
			payload,
			metav1.PatchOptions{FieldManager: openClawDefaultManagedByValue, Force: &force},
		); err != nil {
			return err
		}
	}
	return nil
}

func injectWorkloadTemplateIdentity(
	obj map[string]interface{},
	labels map[string]string,
	annotations map[string]string,
) {
	// 对常见工作负载类型补齐 PodTemplate 元信息，确保 Pod 层也可按 owner/instance 追踪。
	kind := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", obj["kind"])))
	switch kind {
	case "deployment", "statefulset", "daemonset", "replicaset", "job":
		metadata := ensureNestedMetadata(obj, "spec", "template")
		mergeStringValues(metadata, "labels", labels)
		mergeStringValues(metadata, "annotations", annotations)
	case "cronjob":
		metadata := ensureNestedMetadata(obj, "spec", "jobTemplate", "spec", "template")
		mergeStringValues(metadata, "labels", labels)
		mergeStringValues(metadata, "annotations", annotations)
	}
}

func ensureNestedMetadata(root map[string]interface{}, path ...string) map[string]interface{} {
	current := root
	for _, item := range path {
		current = ensureObjectMap(current, item)
	}
	return ensureObjectMap(current, "metadata")
}

func mergeStringValues(root map[string]interface{}, key string, values map[string]string) {
	if len(values) == 0 {
		return
	}
	target := ensureStringMap(root, key)
	for k, v := range values {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			continue
		}
		target[k] = v
	}
	root[key] = target
}

func resolveTemplateResourceGVR(ctx context.Context, disco *discovery.DiscoveryClient, gv schema.GroupVersion, kind string) (schema.GroupVersionResource, bool, error) {
	findInList := func(list *metav1.APIResourceList) (schema.GroupVersionResource, bool, bool) {
		if list == nil {
			return schema.GroupVersionResource{}, false, false
		}
		for i := range list.APIResources {
			resource := list.APIResources[i]
			if strings.Contains(resource.Name, "/") {
				continue
			}
			if !strings.EqualFold(strings.TrimSpace(resource.Kind), strings.TrimSpace(kind)) {
				continue
			}
			if !containsVerb(resource.Verbs, "patch") {
				continue
			}
			return schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource.Name}, resource.Namespaced, true
		}
		return schema.GroupVersionResource{}, false, false
	}

	lists, err := disco.ServerPreferredResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return schema.GroupVersionResource{}, false, err
	}
	expectedGV := gv.String()
	for i := range lists {
		if strings.TrimSpace(lists[i].GroupVersion) != expectedGV {
			continue
		}
		if gvr, namespaced, found := findInList(lists[i]); found {
			return gvr, namespaced, nil
		}
	}

	list, err := disco.ServerResourcesForGroupVersion(expectedGV)
	if err != nil {
		return schema.GroupVersionResource{}, false, fmt.Errorf("%w: resource mapping not found for %s/%s: %v", ErrOpenClawInvalidInput, expectedGV, kind, err)
	}
	if gvr, namespaced, found := findInList(list); found {
		return gvr, namespaced, nil
	}
	return schema.GroupVersionResource{}, false, fmt.Errorf("%w: resource mapping not found for %s/%s", ErrOpenClawInvalidInput, expectedGV, kind)
}

func deleteInstanceResources(ctx context.Context, dc dynamic.Interface, disco *discovery.DiscoveryClient, namespace, instanceName, ownerID string) error {
	instanceSelectors := buildLegacyAwareLabelSelectors(OpenClawInstanceLabelInstance, instanceName)
	ownerSelectors := []string{""}
	if strings.TrimSpace(ownerID) != "" {
		ownerSelectors = buildLegacyAwareLabelSelectors(OpenClawInstanceLabelOwner, ownerID)
	}
	managedBySelectors := buildLegacyAwareLabelSelectors(OpenClawInstanceLabelManagedBy, openClawDefaultManagedByValue)
	typeSelector := fmt.Sprintf("%s=%s", OpenClawInstanceLabelType, openClawInstanceTypeValue)
	selectors := make([]string, 0, len(instanceSelectors)*len(ownerSelectors)*len(managedBySelectors))
	selectorSeen := map[string]struct{}{}
	appendSelector := func(selector string) {
		selector = strings.TrimSpace(selector)
		if selector == "" {
			return
		}
		if _, exists := selectorSeen[selector]; exists {
			return
		}
		selectorSeen[selector] = struct{}{}
		selectors = append(selectors, selector)
	}
	for _, instanceSelector := range instanceSelectors {
		for _, ownerSelector := range ownerSelectors {
			for _, managedBySelector := range managedBySelectors {
				parts := []string{instanceSelector, managedBySelector, typeSelector}
				if ownerSelector != "" {
					parts = append(parts, ownerSelector)
				}
				appendSelector(strings.Join(parts, ","))
				legacyParts := []string{instanceSelector, managedBySelector}
				if ownerSelector != "" {
					legacyParts = append(legacyParts, ownerSelector)
				}
				appendSelector(strings.Join(legacyParts, ","))
			}
		}
	}

	apiResourceLists, err := disco.ServerPreferredNamespacedResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return err
	}
	deletedAny := false
	for _, list := range apiResourceLists {
		gv, parseErr := schema.ParseGroupVersion(list.GroupVersion)
		if parseErr != nil {
			continue
		}
		for _, resource := range list.APIResources {
			if strings.Contains(resource.Name, "/") {
				continue
			}
			if !containsVerb(resource.Verbs, "list") || !containsVerb(resource.Verbs, "delete") {
				continue
			}
			gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource.Name}
			res := dc.Resource(gvr).Namespace(namespace)
			for _, selector := range selectors {
				items, listErr := res.List(ctx, metav1.ListOptions{LabelSelector: selector})
				if listErr != nil {
					if apierrors.IsForbidden(listErr) || apierrors.IsMethodNotSupported(listErr) || apierrors.IsNotFound(listErr) {
						continue
					}
					continue
				}
				for i := range items.Items {
					name := items.Items[i].GetName()
					if name == "" {
						continue
					}
					if delErr := res.Delete(ctx, name, metav1.DeleteOptions{}); delErr != nil && !apierrors.IsNotFound(delErr) {
						return delErr
					}
					deletedAny = true
				}
			}
		}
	}
	if !deletedAny {
		// 兜底清理 Deployment（最小可控资源）
		res := dc.Resource(schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}).Namespace(namespace)
		for _, selector := range selectors {
			if delErr := res.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector}); delErr != nil && !apierrors.IsNotFound(delErr) {
				return delErr
			}
		}
	}
	return nil
}

func containsVerb(verbs []string, expected string) bool {
	for _, verb := range verbs {
		if strings.EqualFold(strings.TrimSpace(verb), expected) {
			return true
		}
	}
	return false
}

func normalizeUsers(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, user := range in {
		value := strings.TrimSpace(user)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func normalizeModels(in OpenClawModelsSpec) OpenClawModelsSpec {
	out := in
	out.Primary = strings.TrimSpace(out.Primary)
	providers := make([]OpenClawProviderSpec, 0, len(in.Providers))
	for i := range in.Providers {
		provider := in.Providers[i]
		provider.Name = strings.TrimSpace(provider.Name)
		if provider.Name == "" {
			continue
		}
		provider.BaseURL = strings.TrimSpace(provider.BaseURL)
		provider.APIType = strings.TrimSpace(provider.APIType)
		provider.ModelIDs = normalizeUsers(provider.ModelIDs)
		if provider.APIKeySecretRef != nil {
			provider.APIKeySecretRef.Name = strings.TrimSpace(provider.APIKeySecretRef.Name)
			provider.APIKeySecretRef.Key = strings.TrimSpace(provider.APIKeySecretRef.Key)
			if provider.APIKeySecretRef.Name == "" || provider.APIKeySecretRef.Key == "" {
				provider.APIKeySecretRef = nil
			}
		}
		providers = append(providers, provider)
	}
	out.Providers = providers
	return out
}

func normalizeAgents(in OpenClawAgentsSpec) OpenClawAgentsSpec {
	out := in
	out.Defaults.ModelPrimary = strings.TrimSpace(out.Defaults.ModelPrimary)
	result := make([]OpenClawAgentSpec, 0, len(in.List))
	for i := range in.List {
		agent := in.List[i]
		agent.ID = strings.TrimSpace(agent.ID)
		if agent.ID == "" {
			continue
		}
		agent.Name = strings.TrimSpace(agent.Name)
		agent.Role = strings.TrimSpace(agent.Role)
		agent.ModelPrimary = strings.TrimSpace(agent.ModelPrimary)
		agent.SandboxProfile = strings.TrimSpace(agent.SandboxProfile)
		agent.Skills = normalizeUsers(agent.Skills)
		agent.Channels = normalizeUsers(agent.Channels)
		result = append(result, agent)
	}
	out.List = result
	return out
}

func normalizeSkills(in OpenClawSkillsSpec) OpenClawSkillsSpec {
	out := in
	out.InstallPolicy = strings.TrimSpace(out.InstallPolicy)
	if out.InstallPolicy == "" {
		out.InstallPolicy = "IfNotPresent"
	}
	refs := make([]OpenClawSkillRef, 0, len(in.Refs))
	seen := map[string]struct{}{}
	for i := range in.Refs {
		ref := in.Refs[i]
		ref.Name = strings.TrimSpace(ref.Name)
		ref.Version = strings.TrimSpace(ref.Version)
		if ref.Name == "" {
			continue
		}
		key := ref.Name + "@" + ref.Version
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		refs = append(refs, ref)
	}
	out.Refs = refs
	return out
}

func normalizeTemplate(in OpenClawTemplateSpec) (OpenClawTemplateSpec, error) {
	out := OpenClawTemplateSpec{
		ConfigMapSelector: strings.TrimSpace(in.ConfigMapSelector),
		Resources:         make([]runtime.RawExtension, 0, len(in.Resources)),
	}
	for i := range in.Resources {
		raw := bytes.TrimSpace(in.Resources[i].Raw)
		if len(raw) == 0 {
			continue
		}
		var obj map[string]interface{}
		if err := json.Unmarshal(raw, &obj); err != nil {
			return OpenClawTemplateSpec{}, fmt.Errorf("%w: template.resources[%d] is not valid JSON object: %v", ErrOpenClawInvalidInput, i, err)
		}
		if _, err := normalizeTemplateObject(obj); err != nil {
			return OpenClawTemplateSpec{}, fmt.Errorf("%w: template.resources[%d]: %v", ErrOpenClawInvalidInput, i, err)
		}
		out.Resources = append(out.Resources, runtime.RawExtension{Raw: raw})
	}
	return out, nil
}

func loadTemplateResourcesFromConfigMaps(ctx context.Context, client *kubernetes.Clientset, namespace, selector string) ([]runtime.RawExtension, error) {
	list, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}
	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].Name < list.Items[j].Name
	})
	out := make([]runtime.RawExtension, 0)
	for i := range list.Items {
		cm := list.Items[i]
		keys := make([]string, 0, len(cm.Data))
		for key := range cm.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			items, parseErr := parseTemplateResourcesYAML(cm.Data[key])
			if parseErr != nil {
				return nil, fmt.Errorf("%w: configmap %s/%s key %s parse failed: %v", ErrOpenClawInvalidInput, namespace, cm.Name, key, parseErr)
			}
			out = append(out, items...)
		}
	}
	return out, nil
}

func parseTemplateResourcesYAML(content string) ([]runtime.RawExtension, error) {
	decoder := utilyaml.NewYAMLOrJSONDecoder(strings.NewReader(content), 4096)
	out := make([]runtime.RawExtension, 0)
	for {
		var obj map[string]interface{}
		if err := decoder.Decode(&obj); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if len(obj) == 0 {
			continue
		}
		raw, err := normalizeTemplateObject(obj)
		if err != nil {
			return nil, err
		}
		out = append(out, runtime.RawExtension{Raw: raw})
	}
	return out, nil
}

func normalizeTemplateObject(obj map[string]interface{}) ([]byte, error) {
	apiVersion, _ := obj["apiVersion"].(string)
	kind, _ := obj["kind"].(string)
	apiVersion = strings.TrimSpace(apiVersion)
	kind = strings.TrimSpace(kind)
	if apiVersion == "" || kind == "" {
		return nil, fmt.Errorf("resource object must include apiVersion and kind")
	}
	raw, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func int32Ptr(v int32) *int32 {
	return &v
}

func generateGatewayToken() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return randomStringFromAlphabet(openClawGatewayTokenLength, alphabet)
}

func injectPlatformTemplateValues(values map[string]interface{}) {
	if values == nil || config.ApplicationConfig == nil {
		return
	}
	control := config.ApplicationConfig.OpenClawControl
	values["ingressEnabled"] = control.IngressEnabled
	values["ingressClassName"] = strings.TrimSpace(control.IngressClassName)
	values["ingressPath"] = strings.TrimSpace(control.IngressPath)
	values["ingressPathType"] = strings.TrimSpace(control.IngressPathType)
	values["ingressTlsEnabled"] = control.IngressTLSEnabled
	values["ingressTlsSecretName"] = strings.TrimSpace(control.IngressTLSSecretName)
}

func resolveTextWithAliases(explicit string, values map[string]interface{}, aliases []string) string {
	if text := strings.TrimSpace(explicit); text != "" {
		return text
	}
	for _, key := range aliases {
		raw, ok := values[key]
		if !ok {
			continue
		}
		text := strings.TrimSpace(fmt.Sprintf("%v", raw))
		if text != "" {
			return text
		}
	}
	return ""
}

func normalizeSecretRef(ref *OpenClawSecretRef) *OpenClawSecretRef {
	if ref == nil {
		return nil
	}
	name := strings.TrimSpace(ref.Name)
	key := strings.TrimSpace(ref.Key)
	if name == "" {
		return nil
	}
	if key == "" {
		key = openClawGatewayTokenDataKey
	}
	return &OpenClawSecretRef{Name: name, Key: key}
}

func inferGatewayTokenSecretRefFromTemplateResources(resources []runtime.RawExtension) *OpenClawSecretRef {
	for i := range resources {
		raw := bytes.TrimSpace(resources[i].Raw)
		if len(raw) == 0 {
			continue
		}
		obj := map[string]interface{}{}
		if err := json.Unmarshal(raw, &obj); err != nil {
			continue
		}
		kind := strings.TrimSpace(fmt.Sprintf("%v", obj["kind"]))
		if !strings.EqualFold(kind, "Deployment") {
			continue
		}
		if ref := inferGatewayTokenSecretRefFromDeployment(obj); ref != nil {
			return ref
		}
	}

	for i := range resources {
		raw := bytes.TrimSpace(resources[i].Raw)
		if len(raw) == 0 {
			continue
		}
		obj := map[string]interface{}{}
		if err := json.Unmarshal(raw, &obj); err != nil {
			continue
		}
		kind := strings.TrimSpace(fmt.Sprintf("%v", obj["kind"]))
		if !strings.EqualFold(kind, "Secret") {
			continue
		}
		meta := ensureObjectMap(obj, "metadata")
		name := strings.TrimSpace(fmt.Sprintf("%v", meta["name"]))
		if name == "" {
			continue
		}
		if hasGatewayTokenInSecret(obj) {
			return &OpenClawSecretRef{Name: name, Key: openClawGatewayTokenDataKey}
		}
	}
	return nil
}

func inferGatewayTokenSecretRefFromDeployment(obj map[string]interface{}) *OpenClawSecretRef {
	containersRaw, found, err := unstructured.NestedSlice(obj, "spec", "template", "spec", "containers")
	if err != nil || !found {
		return nil
	}
	for _, item := range containersRaw {
		container, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		envList, found, err := unstructured.NestedSlice(container, "env")
		if err != nil || !found {
			continue
		}
		for _, envRaw := range envList {
			env, ok := envRaw.(map[string]interface{})
			if !ok {
				continue
			}
			key := strings.TrimSpace(fmt.Sprintf("%v", env["name"]))
			if key != openClawGatewayTokenDataKey {
				continue
			}
			refName, found, err := unstructured.NestedString(env, "valueFrom", "secretKeyRef", "name")
			if err != nil || !found || strings.TrimSpace(refName) == "" {
				continue
			}
			refKey, found, err := unstructured.NestedString(env, "valueFrom", "secretKeyRef", "key")
			if err != nil || !found || strings.TrimSpace(refKey) == "" {
				refKey = openClawGatewayTokenDataKey
			}
			return &OpenClawSecretRef{Name: strings.TrimSpace(refName), Key: strings.TrimSpace(refKey)}
		}
	}
	return nil
}

func hasGatewayTokenInSecret(obj map[string]interface{}) bool {
	if data, ok := obj["data"].(map[string]interface{}); ok {
		if _, found := data[openClawGatewayTokenDataKey]; found {
			return true
		}
	}
	if data, ok := obj["stringData"].(map[string]interface{}); ok {
		if _, found := data[openClawGatewayTokenDataKey]; found {
			return true
		}
	}
	return false
}

func enrichInstanceGatewayTokens(ctx context.Context, kc *kubernetes.Clientset, instances []OpenClawInstance, requesterID string) {
	if kc == nil || len(instances) == 0 {
		return
	}
	requesterID = strings.TrimSpace(requesterID)
	cache := map[string]string{}
	for i := range instances {
		ownerID := strings.TrimSpace(instances[i].OwnerID)
		if requesterID == "" || ownerID == "" || requesterID != ownerID {
			instances[i].Status.GatewayToken = ""
			continue
		}
		ref := instanceGatewayTokenSecretRef(instances[i])
		if ref == nil {
			instances[i].Status.GatewayToken = ""
			continue
		}
		cacheKey := fmt.Sprintf("%s/%s/%s", instances[i].Namespace, ref.Name, ref.Key)
		token, ok := cache[cacheKey]
		if !ok {
			resolved, err := readGatewayTokenFromSecret(ctx, kc, instances[i].Namespace, *ref)
			if err != nil {
				config.Logger.Warnf("openclaw resolve gateway token failed namespace=%s instance=%s secret=%s key=%s err=%v",
					instances[i].Namespace, instances[i].Name, ref.Name, ref.Key, err)
				cache[cacheKey] = ""
				continue
			}
			token = strings.TrimSpace(resolved)
			cache[cacheKey] = token
		}
		if token == "" {
			continue
		}
		instances[i].Status.GatewayToken = token
	}
}

func instanceGatewayTokenSecretRef(instance OpenClawInstance) *OpenClawSecretRef {
	if ref := normalizeSecretRef(instance.Gateway.TokenSecretRef); ref != nil {
		return ref
	}
	name := ""
	key := ""
	if instance.Annotations != nil {
		name = readMapValueWithLegacy(instance.Annotations, openClawGatewayTokenSecretName)
		key = readMapValueWithLegacy(instance.Annotations, openClawGatewayTokenSecretKey)
	}
	if name == "" {
		return nil
	}
	if key == "" {
		key = openClawGatewayTokenDataKey
	}
	return &OpenClawSecretRef{Name: name, Key: key}
}

func readGatewayTokenFromSecret(ctx context.Context, kc *kubernetes.Clientset, namespace string, ref OpenClawSecretRef) (string, error) {
	normalized := normalizeSecretRef(&ref)
	if normalized == nil {
		return "", fmt.Errorf("invalid gateway token secret ref")
	}
	ref = *normalized
	secret, err := kc.CoreV1().Secrets(namespace).Get(ctx, ref.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	raw, ok := secret.Data[ref.Key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret %s/%s", ref.Key, namespace, ref.Name)
	}
	return strings.TrimSpace(string(raw)), nil
}

type openClawConfigModelHints struct {
	Primary   string
	Providers []OpenClawProviderSpec
}

type openClawRuntimeModelConfig struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	API  string `json:"api,omitempty"`
}

type openClawRuntimeProviderConfig struct {
	BaseURL string                       `json:"baseUrl,omitempty"`
	APIType string                       `json:"apiType,omitempty"`
	Type    string                       `json:"type,omitempty"`
	Models  []openClawRuntimeModelConfig `json:"models,omitempty"`
}

type openClawRuntimeConfig struct {
	Models struct {
		Primary   string                     `json:"primary,omitempty"`
		Providers map[string]json.RawMessage `json:"providers,omitempty"`
	} `json:"models,omitempty"`
	Agents struct {
		Defaults struct {
			Model struct {
				Primary string `json:"primary,omitempty"`
			} `json:"model,omitempty"`
		} `json:"defaults,omitempty"`
	} `json:"agents,omitempty"`
}

func enrichInstanceModelProviders(ctx context.Context, kc *kubernetes.Clientset, instances []OpenClawInstance) {
	if kc == nil || len(instances) == 0 {
		return
	}
	cache := map[string]openClawConfigModelHints{}
	for i := range instances {
		namespace := strings.TrimSpace(instances[i].Namespace)
		name := strings.TrimSpace(instances[i].Name)
		if namespace == "" || name == "" {
			continue
		}
		secretName := fmt.Sprintf("%s-openclaw-config", name)
		cacheKey := fmt.Sprintf("%s/%s", namespace, secretName)
		hints, ok := cache[cacheKey]
		if !ok {
			resolved, err := readModelHintsFromConfigSecret(ctx, kc, namespace, secretName)
			if err != nil {
				cache[cacheKey] = openClawConfigModelHints{}
				continue
			}
			hints = resolved
			cache[cacheKey] = hints
		}
		if hints.Primary == "" && len(hints.Providers) == 0 {
			continue
		}
		if strings.TrimSpace(instances[i].Models.Primary) == "" && hints.Primary != "" {
			instances[i].Models.Primary = hints.Primary
		}
		if len(hints.Providers) == 0 {
			continue
		}
		if len(instances[i].Models.Providers) == 0 {
			instances[i].Models.Providers = hints.Providers
			continue
		}
		existing := map[string]int{}
		for idx := range instances[i].Models.Providers {
			key := strings.TrimSpace(instances[i].Models.Providers[idx].Name)
			if key == "" {
				continue
			}
			existing[key] = idx
		}
		for _, provider := range hints.Providers {
			key := strings.TrimSpace(provider.Name)
			if key == "" {
				continue
			}
			if idx, ok := existing[key]; ok {
				if strings.TrimSpace(instances[i].Models.Providers[idx].BaseURL) == "" {
					instances[i].Models.Providers[idx].BaseURL = provider.BaseURL
				}
				if strings.TrimSpace(instances[i].Models.Providers[idx].APIType) == "" {
					instances[i].Models.Providers[idx].APIType = provider.APIType
				}
				if len(instances[i].Models.Providers[idx].ModelIDs) == 0 && len(provider.ModelIDs) > 0 {
					instances[i].Models.Providers[idx].ModelIDs = provider.ModelIDs
				}
				continue
			}
			instances[i].Models.Providers = append(instances[i].Models.Providers, provider)
		}
	}
}

func readModelHintsFromConfigSecret(ctx context.Context, kc *kubernetes.Clientset, namespace, secretName string) (openClawConfigModelHints, error) {
	secret, err := kc.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return openClawConfigModelHints{}, err
	}
	raw, ok := secret.Data["openclaw.json"]
	if !ok || len(bytes.TrimSpace(raw)) == 0 {
		return openClawConfigModelHints{}, fmt.Errorf("openclaw.json not found in secret %s/%s", namespace, secretName)
	}
	cfg := openClawRuntimeConfig{}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return openClawConfigModelHints{}, err
	}
	primary := strings.TrimSpace(cfg.Models.Primary)
	if primary == "" {
		primary = strings.TrimSpace(cfg.Agents.Defaults.Model.Primary)
	}
	providers := make([]OpenClawProviderSpec, 0, len(cfg.Models.Providers))
	for providerName, providerRaw := range cfg.Models.Providers {
		trimmed := strings.TrimSpace(providerName)
		if trimmed == "" {
			continue
		}
		provider := OpenClawProviderSpec{Name: trimmed}
		parsedProvider := openClawRuntimeProviderConfig{}
		if err := json.Unmarshal(providerRaw, &parsedProvider); err == nil {
			provider.BaseURL = strings.TrimSpace(parsedProvider.BaseURL)
			provider.APIType = strings.TrimSpace(parsedProvider.APIType)
			if provider.APIType == "" {
				provider.APIType = strings.TrimSpace(parsedProvider.Type)
			}
			modelIDs := make([]string, 0, len(parsedProvider.Models))
			apiTypes := map[string]struct{}{}
			for _, model := range parsedProvider.Models {
				modelID := strings.TrimSpace(model.ID)
				if modelID == "" {
					modelID = strings.TrimSpace(model.Name)
				}
				if modelID != "" {
					modelIDs = append(modelIDs, modelID)
				}
				modelAPI := strings.TrimSpace(model.API)
				if modelAPI != "" {
					apiTypes[modelAPI] = struct{}{}
				}
			}
			modelIDs = normalizeUsers(modelIDs)
			provider.ModelIDs = modelIDs
			if provider.APIType == "" && len(apiTypes) == 1 {
				for modelAPI := range apiTypes {
					provider.APIType = strings.TrimSpace(modelAPI)
				}
			}
		}
		providers = append(providers, provider)
	}
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Name < providers[j].Name
	})
	return openClawConfigModelHints{
		Primary:   primary,
		Providers: providers,
	}, nil
}

func buildPreviewEndpoint(namespace, instanceName, baseDomain string) string {
	ns := strings.TrimSpace(namespace)
	name := strings.TrimSpace(instanceName)
	domain := strings.TrimSpace(baseDomain)
	if ns == "" || name == "" || domain == "" {
		return ""
	}
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.Trim(domain, "/")
	if domain == "" {
		return ""
	}
	return fmt.Sprintf("https://%s-%s.%s", name, ns, domain)
}

func extractEndpointHost(rawEndpoint string) string {
	endpoint := strings.TrimSpace(rawEndpoint)
	if endpoint == "" {
		return ""
	}
	// 兼容仅传入 host 的场景。
	if !strings.HasPrefix(strings.ToLower(endpoint), "http://") && !strings.HasPrefix(strings.ToLower(endpoint), "https://") {
		endpoint = "https://" + endpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Host)
}

func generateInstanceName(ctx context.Context, kc *kubernetes.Clientset, namespace string) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	for i := 0; i < 20; i++ {
		suffix, err := randomStringFromAlphabet(openClawGeneratedNameLength, alphabet)
		if err != nil {
			return "", err
		}
		name := openClawGeneratedNamePrefix + suffix
		if errs := validation.IsDNS1123Subdomain(name); len(errs) > 0 {
			continue
		}
		existing, listErr := listInstanceDeployments(ctx, kc, namespace, name)
		if listErr != nil {
			return "", listErr
		}
		if len(existing) == 0 {
			return name, nil
		}
	}
	return "", fmt.Errorf("%w: failed to generate unique instance name", ErrOpenClawInvalidInput)
}

func randomStringFromAlphabet(length int, alphabet string) (string, error) {
	if length <= 0 || len(alphabet) == 0 {
		return "", fmt.Errorf("%w: invalid random string arguments", ErrOpenClawInvalidInput)
	}
	buf := make([]byte, length)
	if _, err := cryptorand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, length)
	for i := 0; i < length; i++ {
		out[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(out), nil
}
