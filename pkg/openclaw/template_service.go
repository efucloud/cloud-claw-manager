package openclaw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/efucloud/cloud-claw-manager/pkg/config"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	OpenClawTemplateLabelKey        = "efucloud.com/openclaw.template"
	OpenClawTemplateLabelValue      = "true"
	openClawTemplateDataSchemaKey   = "schema"
	openClawTemplateDataDefaultsKey = "defaults"
	openClawTemplateDataBodyKey     = "templates"
	openClawTemplateDescAnnotation  = "efucloud.com/openclaw.template.description"

	openClawResourceLabelOwner    = "efucloud.com/openclaw.owner"
	openClawResourceLabelInstance = "efucloud.com/openclaw.instance"
)

var (
	templateNamespaceOnce sync.Once
	templateNamespace     string
)

type OpenClawTemplateService struct{}

func (s OpenClawTemplateService) ListTemplates(ctx context.Context) (OpenClawTemplateListResponse, error) {
	client, err := s.templateClient(ctx)
	if err != nil {
		return OpenClawTemplateListResponse{}, err
	}
	list, err := client.List(ctx, metav1.ListOptions{LabelSelector: OpenClawTemplateLabelKey + "=" + OpenClawTemplateLabelValue})
	if err != nil {
		return OpenClawTemplateListResponse{}, err
	}
	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].Name < list.Items[j].Name
	})
	out := make([]OpenClawTemplate, 0, len(list.Items))
	for i := range list.Items {
		item, parseErr := parseTemplateConfigMap(list.Items[i])
		if parseErr != nil {
			config.Logger.Warnf("skip invalid openclaw template configmap %s/%s: %v", list.Items[i].Namespace, list.Items[i].Name, parseErr)
			continue
		}
		item.Templates = ""
		out = append(out, item)
	}
	return OpenClawTemplateListResponse{Total: uint(len(out)), Data: out}, nil
}

func (s OpenClawTemplateService) GetTemplate(ctx context.Context, name string) (OpenClawTemplate, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return OpenClawTemplate{}, fmt.Errorf("%w: template name is required", ErrOpenClawInvalidInput)
	}
	client, err := s.templateClient(ctx)
	if err != nil {
		return OpenClawTemplate{}, err
	}
	cm, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return OpenClawTemplate{}, ErrOpenClawNotFound
		}
		return OpenClawTemplate{}, err
	}
	return parseTemplateConfigMap(*cm)
}

func (s OpenClawTemplateService) CreateTemplate(ctx context.Context, requesterEmail string, req OpenClawTemplateUpsertRequest) (OpenClawTemplate, error) {
	if !IsTemplateAdmin(requesterEmail) {
		return OpenClawTemplate{}, ErrOpenClawForbidden
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return OpenClawTemplate{}, fmt.Errorf("%w: template name is required", ErrOpenClawInvalidInput)
	}
	if errs := validation.IsDNS1123Subdomain(req.Name); len(errs) > 0 {
		return OpenClawTemplate{}, fmt.Errorf("%w: invalid template name: %s", ErrOpenClawInvalidInput, strings.Join(errs, ","))
	}
	normalized, err := normalizeTemplateUpsert(req)
	if err != nil {
		return OpenClawTemplate{}, err
	}
	client, err := s.templateClient(ctx)
	if err != nil {
		return OpenClawTemplate{}, err
	}
	if _, err = client.Get(ctx, normalized.Name, metav1.GetOptions{}); err == nil {
		return OpenClawTemplate{}, fmt.Errorf("%w: template already exists", ErrOpenClawInvalidInput)
	} else if !apierrors.IsNotFound(err) {
		return OpenClawTemplate{}, err
	}
	cm, err := templateToConfigMap(resolveTemplateNamespace(), normalized)
	if err != nil {
		return OpenClawTemplate{}, err
	}
	created, err := client.Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return OpenClawTemplate{}, err
	}
	return parseTemplateConfigMap(*created)
}

func (s OpenClawTemplateService) UpdateTemplate(ctx context.Context, requesterEmail, name string, req OpenClawTemplateUpsertRequest) (OpenClawTemplate, error) {
	if !IsTemplateAdmin(requesterEmail) {
		return OpenClawTemplate{}, ErrOpenClawForbidden
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return OpenClawTemplate{}, fmt.Errorf("%w: template name is required", ErrOpenClawInvalidInput)
	}
	if strings.TrimSpace(req.Name) != "" && strings.TrimSpace(req.Name) != name {
		return OpenClawTemplate{}, fmt.Errorf("%w: template name in body mismatches path", ErrOpenClawInvalidInput)
	}
	req.Name = name
	normalized, err := normalizeTemplateUpsert(req)
	if err != nil {
		return OpenClawTemplate{}, err
	}
	client, err := s.templateClient(ctx)
	if err != nil {
		return OpenClawTemplate{}, err
	}
	existing, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return OpenClawTemplate{}, ErrOpenClawNotFound
		}
		return OpenClawTemplate{}, err
	}
	next, err := templateToConfigMap(existing.Namespace, normalized)
	if err != nil {
		return OpenClawTemplate{}, err
	}
	next.ResourceVersion = existing.ResourceVersion
	updated, err := client.Update(ctx, next, metav1.UpdateOptions{})
	if err != nil {
		return OpenClawTemplate{}, err
	}
	return parseTemplateConfigMap(*updated)
}

func (s OpenClawTemplateService) RenderTemplate(ctx context.Context, name string, input map[string]interface{}) ([]runtime.RawExtension, error) {
	tpl, err := s.GetTemplate(ctx, name)
	if err != nil {
		return nil, err
	}
	values := mergeTemplateValues(tpl.Defaults, input)
	if err = validateTemplateValues(tpl.Schema, values); err != nil {
		return nil, err
	}
	rendered, err := renderTemplateBody(tpl.Templates, values)
	if err != nil {
		return nil, fmt.Errorf("%w: render template %s failed: %v", ErrOpenClawInvalidInput, name, err)
	}
	resources, err := parseTemplateResourcesYAML(rendered)
	if err != nil {
		return nil, fmt.Errorf("%w: parse rendered template %s failed: %v", ErrOpenClawInvalidInput, name, err)
	}
	if len(resources) == 0 {
		return nil, fmt.Errorf("%w: template %s rendered no resources", ErrOpenClawInvalidInput, name)
	}
	return resources, nil
}

func IsTemplateAdmin(email string) bool {
	candidate := strings.ToLower(strings.TrimSpace(email))
	if candidate == "" {
		return false
	}
	for _, item := range config.ApplicationConfig.OpenClawControl.AdminEmails {
		if strings.ToLower(strings.TrimSpace(item)) == candidate {
			return true
		}
	}
	return false
}

func InjectOpenClawOwnershipLabels(resources []runtime.RawExtension, owner, instance string) ([]runtime.RawExtension, error) {
	owner = strings.TrimSpace(owner)
	instance = strings.TrimSpace(instance)
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
		meta := ensureObjectMap(obj, "metadata")
		labels := ensureStringMap(meta, "labels")
		if owner != "" {
			labels[openClawResourceLabelOwner] = owner
		}
		if instance != "" {
			labels[openClawResourceLabelInstance] = instance
		}
		meta["labels"] = labels
		obj["metadata"] = meta
		nextRaw, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		out = append(out, runtime.RawExtension{Raw: nextRaw})
	}
	return out, nil
}

func parseTemplateConfigMap(cm corev1.ConfigMap) (OpenClawTemplate, error) {
	schema, err := parseJSONMap(cm.Data[openClawTemplateDataSchemaKey], true)
	if err != nil {
		return OpenClawTemplate{}, fmt.Errorf("%w: schema parse failed: %v", ErrOpenClawInvalidInput, err)
	}
	defaults, err := parseJSONMap(cm.Data[openClawTemplateDataDefaultsKey], false)
	if err != nil {
		return OpenClawTemplate{}, fmt.Errorf("%w: defaults parse failed: %v", ErrOpenClawInvalidInput, err)
	}
	templates := strings.TrimSpace(cm.Data[openClawTemplateDataBodyKey])
	if templates == "" {
		return OpenClawTemplate{}, fmt.Errorf("%w: templates is required", ErrOpenClawInvalidInput)
	}
	updatedAt := cm.CreationTimestamp.Time
	for i := range cm.ManagedFields {
		if cm.ManagedFields[i].Time != nil && cm.ManagedFields[i].Time.After(updatedAt) {
			updatedAt = cm.ManagedFields[i].Time.Time
		}
	}
	return OpenClawTemplate{
		Name:        cm.Name,
		Namespace:   cm.Namespace,
		Description: strings.TrimSpace(cm.Annotations[openClawTemplateDescAnnotation]),
		Schema:      schema,
		Defaults:    defaults,
		Templates:   templates,
		CreatedAt:   cm.CreationTimestamp.Time,
		UpdatedAt:   updatedAt,
	}, nil
}

func normalizeTemplateUpsert(req OpenClawTemplateUpsertRequest) (OpenClawTemplateUpsertRequest, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Templates = strings.TrimSpace(req.Templates)
	if req.Templates == "" {
		return OpenClawTemplateUpsertRequest{}, fmt.Errorf("%w: templates is required", ErrOpenClawInvalidInput)
	}
	if _, err := template.New("openclaw-template").Parse(req.Templates); err != nil {
		return OpenClawTemplateUpsertRequest{}, fmt.Errorf("%w: templates parse failed: %v", ErrOpenClawInvalidInput, err)
	}
	if req.Schema == nil {
		req.Schema = map[string]interface{}{"type": "object"}
	}
	if req.Defaults == nil {
		req.Defaults = map[string]interface{}{}
	}
	if err := validateTemplateValues(req.Schema, req.Defaults); err != nil {
		return OpenClawTemplateUpsertRequest{}, err
	}
	return req, nil
}

func templateToConfigMap(namespace string, req OpenClawTemplateUpsertRequest) (*corev1.ConfigMap, error) {
	schemaJSON, err := marshalTemplateMap(req.Schema)
	if err != nil {
		return nil, fmt.Errorf("%w: serialize schema failed: %v", ErrOpenClawInvalidInput, err)
	}
	defaultsJSON, err := marshalTemplateMap(req.Defaults)
	if err != nil {
		return nil, fmt.Errorf("%w: serialize defaults failed: %v", ErrOpenClawInvalidInput, err)
	}
	annotations := map[string]string{}
	if req.Description != "" {
		annotations[openClawTemplateDescAnnotation] = req.Description
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   namespace,
			Name:        req.Name,
			Labels:      map[string]string{OpenClawTemplateLabelKey: OpenClawTemplateLabelValue},
			Annotations: annotations,
		},
		Data: map[string]string{
			openClawTemplateDataSchemaKey:   schemaJSON,
			openClawTemplateDataDefaultsKey: defaultsJSON,
			openClawTemplateDataBodyKey:     req.Templates,
		},
	}, nil
}

func marshalTemplateMap(data map[string]interface{}) (string, error) {
	if data == nil {
		data = map[string]interface{}{}
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func mergeTemplateValues(defaults, input map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range defaults {
		out[k] = v
	}
	for k, v := range input {
		out[k] = v
	}
	return out
}

func validateTemplateValues(schema map[string]interface{}, values map[string]interface{}) error {
	if len(schema) == 0 {
		return nil
	}
	required, _ := schema["required"].([]interface{})
	for _, item := range required {
		key, _ := item.(string)
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := values[key]; !ok {
			return fmt.Errorf("%w: required template value %q is missing", ErrOpenClawInvalidInput, key)
		}
	}
	properties, _ := schema["properties"].(map[string]interface{})
	for key, raw := range properties {
		prop, _ := raw.(map[string]interface{})
		typeName, _ := prop["type"].(string)
		typeName = strings.TrimSpace(typeName)
		if typeName == "" {
			continue
		}
		val, ok := values[key]
		if !ok {
			continue
		}
		if !matchSchemaType(typeName, val) {
			return fmt.Errorf("%w: template value %q does not match schema type %q", ErrOpenClawInvalidInput, key, typeName)
		}
	}
	return nil
}

func matchSchemaType(typeName string, val interface{}) bool {
	switch typeName {
	case "string":
		_, ok := val.(string)
		return ok
	case "number":
		switch val.(type) {
		case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return true
		default:
			return false
		}
	case "integer":
		switch v := val.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return true
		case float64:
			return v == float64(int64(v))
		case float32:
			return v == float32(int64(v))
		default:
			return false
		}
	case "boolean":
		_, ok := val.(bool)
		return ok
	case "object":
		_, ok := val.(map[string]interface{})
		return ok
	case "array":
		_, ok := val.([]interface{})
		return ok
	default:
		return true
	}
}

func renderTemplateBody(content string, values map[string]interface{}) (string, error) {
	tpl, err := template.New("openclaw-template").Option("missingkey=error").Parse(content)
	if err != nil {
		return "", err
	}
	data := map[string]interface{}{"Values": values}
	for key, val := range values {
		data[key] = val
	}
	buffer := bytes.NewBuffer(nil)
	if err = tpl.Execute(buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func parseJSONMap(raw string, required bool) (map[string]interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required {
			return nil, fmt.Errorf("value is required")
		}
		return map[string]interface{}{}, nil
	}
	out := map[string]interface{}{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func ensureObjectMap(root map[string]interface{}, key string) map[string]interface{} {
	raw, ok := root[key]
	if !ok {
		obj := map[string]interface{}{}
		root[key] = obj
		return obj
	}
	obj, ok := raw.(map[string]interface{})
	if ok {
		return obj
	}
	obj = map[string]interface{}{}
	root[key] = obj
	return obj
}

func ensureStringMap(root map[string]interface{}, key string) map[string]string {
	raw, ok := root[key]
	if !ok {
		result := map[string]string{}
		root[key] = result
		return result
	}
	if result, ok := raw.(map[string]string); ok {
		return result
	}
	if anyMap, ok := raw.(map[string]interface{}); ok {
		result := map[string]string{}
		for k, v := range anyMap {
			result[k] = fmt.Sprintf("%v", v)
		}
		root[key] = result
		return result
	}
	result := map[string]string{}
	root[key] = result
	return result
}

func resolveTemplateNamespace() string {
	templateNamespaceOnce.Do(func() {
		if strings.TrimSpace(config.RunNamespace) != "" {
			templateNamespace = strings.TrimSpace(config.RunNamespace)
			return
		}
		templateNamespace = "openclaw"
		config.RunNamespace = templateNamespace
	})
	return templateNamespace
}

func (s OpenClawTemplateService) templateClient(ctx context.Context) (coreclientv1.ConfigMapInterface, error) {
	kc, _, _, err := defaultOpenClawClients.get(ctx)
	if err != nil {
		return nil, err
	}
	return kc.CoreV1().ConfigMaps(resolveTemplateNamespace()), nil
}
