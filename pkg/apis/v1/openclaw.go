package v1

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/efucloud/cloud-claw-manager/pkg/apis/filters"
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/cloud-claw-manager/pkg/dtos"
	"github.com/efucloud/cloud-claw-manager/pkg/openclaw"
	"github.com/efucloud/common"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
)

type OpenClawResource struct{}

func (r OpenClawResource) AddWebService(ws *restful.WebService) {
	apiInfo := common.ApiInfo{}
	apiInfo.Tag = "openclaw"
	apiInfo.Description = "OpenClaw 管理"
	common.RegisterApiInfo(apiInfo)

	ws.Route(ws.GET(config.APIPrefix+"/openclaw/dashboard").
		Doc("获取当前用户可访问 OpenClaw 实例概览").
		Notes("基于 OpenClaw 模板下发后的 Kubernetes 资源聚合实例信息").
		Param(ws.QueryParameter("namespace", "按 namespace 过滤（可选）")).
		Param(ws.QueryParameter("includeGatewayToken", "是否返回实例 gateway token（可选，默认 true）")).
		Returns(http.StatusOK, "请求成功", openclaw.OpenClawDashboardResponse{}).
		Returns(http.StatusUnauthorized, "未认证", dtos.ResponseError{}).
		Returns(http.StatusInternalServerError, "内部处理逻辑错误", dtos.ResponseError{}).
		To(r.dashboard).
		Filter(filters.I18n).Filter(filters.Log).Filter(filters.Auth).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "openclawDashboard"))

	ws.Route(ws.GET(config.APIPrefix+"/openclaw/instances").
		Doc("获取当前用户可访问 OpenClaw 实例列表").
		Param(ws.QueryParameter("namespace", "按 namespace 过滤（可选）")).
		Returns(http.StatusOK, "请求成功", openclaw.OpenClawInstanceListResponse{}).
		To(r.listInstances).
		Filter(filters.I18n).Filter(filters.Log).Filter(filters.Auth).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "openclawInstances"))

	ws.Route(ws.POST(config.APIPrefix+"/openclaw/instances").
		Doc("创建 OpenClaw 实例（基于模板渲染并下发 YAML 资源）").
		Returns(http.StatusOK, "请求成功", openclaw.OpenClawInstance{}).
		To(r.createInstance).
		Filter(filters.I18n).Filter(filters.Log).Filter(filters.Auth).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "openclawCreateInstance"))

	ws.Route(ws.GET(config.APIPrefix+"/openclaw/instances/{namespace}/{name}").
		Doc("获取单个 OpenClaw 实例详情").
		Param(ws.PathParameter("namespace", "实例命名空间")).
		Param(ws.PathParameter("name", "实例名称")).
		Returns(http.StatusOK, "请求成功", openclaw.OpenClawInstance{}).
		To(r.getInstance).
		Filter(filters.I18n).Filter(filters.Log).Filter(filters.Auth).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "openclawGetInstance"))

	ws.Route(ws.DELETE(config.APIPrefix+"/openclaw/instances/{namespace}/{name}").
		Doc("删除 OpenClaw 实例（删除模板下发的实例资源）").
		Param(ws.PathParameter("namespace", "实例命名空间")).
		Param(ws.PathParameter("name", "实例名称")).
		Returns(http.StatusOK, "请求成功", dtos.ResponseError{}).
		To(r.deleteInstance).
		Filter(filters.I18n).Filter(filters.Log).Filter(filters.Auth).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "openclawDeleteInstance"))

	ws.Route(ws.POST(config.APIPrefix+"/openclaw/instances/{namespace}/{name}/{action}").
		Doc("控制 OpenClaw 实例动作：start/stop/restart").
		Param(ws.PathParameter("namespace", "实例命名空间")).
		Param(ws.PathParameter("name", "实例名称")).
		Param(ws.PathParameter("action", "动作：start/stop/restart")).
		Returns(http.StatusOK, "请求成功", openclaw.OpenClawInstanceActionResponse{}).
		To(r.controlInstance).
		Filter(filters.I18n).Filter(filters.Log).Filter(filters.Auth).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "openclawControlInstance"))

	ws.Route(ws.GET(config.APIPrefix+"/openclaw/templates").
		Doc("获取 OpenClaw 模板列表（基于 ConfigMap）").
		Returns(http.StatusOK, "请求成功", openclaw.OpenClawTemplateListResponse{}).
		To(r.listTemplates).
		Filter(filters.I18n).Filter(filters.Log).Filter(filters.Auth).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "openclawTemplates"))

	ws.Route(ws.GET(config.APIPrefix+"/openclaw/templates/{name}").
		Doc("获取单个 OpenClaw 模板详情（包含 schema/defaults/templates）").
		Param(ws.PathParameter("name", "模板名称（ConfigMap 名）")).
		Returns(http.StatusOK, "请求成功", openclaw.OpenClawTemplate{}).
		To(r.getTemplate).
		Filter(filters.I18n).Filter(filters.Log).Filter(filters.Auth).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "openclawGetTemplate"))

	ws.Route(ws.POST(config.APIPrefix+"/openclaw/templates").
		Doc("创建 OpenClaw 模板（管理员）").
		Reads(openclaw.OpenClawTemplateUpsertRequest{}).
		Returns(http.StatusOK, "请求成功", openclaw.OpenClawTemplate{}).
		To(r.createTemplate).
		Filter(filters.I18n).Filter(filters.Log).Filter(filters.Auth).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "openclawCreateTemplate"))

	ws.Route(ws.PUT(config.APIPrefix+"/openclaw/templates/{name}").
		Doc("更新 OpenClaw 模板（管理员）").
		Param(ws.PathParameter("name", "模板名称（ConfigMap 名）")).
		Reads(openclaw.OpenClawTemplateUpsertRequest{}).
		Returns(http.StatusOK, "请求成功", openclaw.OpenClawTemplate{}).
		To(r.updateTemplate).
		Filter(filters.I18n).Filter(filters.Log).Filter(filters.Auth).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "openclawUpdateTemplate"))
}

func (r OpenClawResource) dashboard(req *restful.Request, resp *restful.Response) {
	requester := requesterID(req)
	if requester == "" {
		resp.WriteErrorString(http.StatusUnauthorized, "unauthorized")
		return
	}
	list, err := (openclaw.OpenClawInstanceService{}).ListInstances(req.Request.Context(), requester, req.QueryParameter("namespace"))
	if err != nil {
		writeOpenClawError(resp, err)
		return
	}
	includeGatewayToken := parseBoolQueryWithDefault(req.QueryParameter("includeGatewayToken"), true)
	common.ResponseSuccess(resp, buildDashboardFromInstances(list.Data, includeGatewayToken))
}

func (r OpenClawResource) listInstances(req *restful.Request, resp *restful.Response) {
	requester := requesterID(req)
	if requester == "" {
		resp.WriteErrorString(http.StatusUnauthorized, "unauthorized")
		return
	}
	data, err := (openclaw.OpenClawInstanceService{}).ListInstances(req.Request.Context(), requester, req.QueryParameter("namespace"))
	if err != nil {
		writeOpenClawError(resp, err)
		return
	}
	common.ResponseSuccess(resp, data)
}

func (r OpenClawResource) createInstance(req *restful.Request, resp *restful.Response) {
	requester := requesterID(req)
	if requester == "" {
		resp.WriteErrorString(http.StatusUnauthorized, "unauthorized")
		return
	}
	var body openclaw.OpenClawInstanceCreateRequest
	if err := req.ReadEntity(&body); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	data, err := (openclaw.OpenClawInstanceService{}).CreateInstance(
		req.Request.Context(),
		requester,
		requesterOwnerMeta(req),
		body,
	)
	if err != nil {
		writeOpenClawError(resp, err)
		return
	}
	common.ResponseSuccess(resp, data)
}

func (r OpenClawResource) getInstance(req *restful.Request, resp *restful.Response) {
	requester := requesterID(req)
	if requester == "" {
		resp.WriteErrorString(http.StatusUnauthorized, "unauthorized")
		return
	}
	data, err := (openclaw.OpenClawInstanceService{}).GetInstance(
		req.Request.Context(),
		requester,
		req.PathParameter("namespace"),
		req.PathParameter("name"),
	)
	if err != nil {
		writeOpenClawError(resp, err)
		return
	}
	common.ResponseSuccess(resp, data)
}

func (r OpenClawResource) deleteInstance(req *restful.Request, resp *restful.Response) {
	requester := requesterID(req)
	if requester == "" {
		resp.WriteErrorString(http.StatusUnauthorized, "unauthorized")
		return
	}
	err := (openclaw.OpenClawInstanceService{}).DeleteInstance(
		req.Request.Context(),
		requester,
		req.PathParameter("namespace"),
		req.PathParameter("name"),
	)
	if err != nil {
		writeOpenClawError(resp, err)
		return
	}
	common.ResponseSuccess(resp, map[string]string{"message": "deleted"})
}

func (r OpenClawResource) controlInstance(req *restful.Request, resp *restful.Response) {
	requester := requesterID(req)
	if requester == "" {
		resp.WriteErrorString(http.StatusUnauthorized, "unauthorized")
		return
	}
	action := strings.ToLower(strings.TrimSpace(req.PathParameter("action")))
	data, err := (openclaw.OpenClawInstanceService{}).ControlInstance(
		req.Request.Context(),
		requester,
		req.PathParameter("namespace"),
		req.PathParameter("name"),
		action,
	)
	if err != nil {
		writeOpenClawError(resp, err)
		return
	}
	common.ResponseSuccess(resp, openclaw.OpenClawInstanceActionResponse{
		Action:  action,
		Message: "ok",
		Data:    data,
	})
}

func (r OpenClawResource) listTemplates(req *restful.Request, resp *restful.Response) {
	data, err := (openclaw.OpenClawTemplateService{}).ListTemplates(req.Request.Context())
	if err != nil {
		writeOpenClawError(resp, err)
		return
	}
	common.ResponseSuccess(resp, data)
}

func (r OpenClawResource) getTemplate(req *restful.Request, resp *restful.Response) {
	data, err := (openclaw.OpenClawTemplateService{}).GetTemplate(req.Request.Context(), req.PathParameter("name"))
	if err != nil {
		writeOpenClawError(resp, err)
		return
	}
	common.ResponseSuccess(resp, data)
}

func (r OpenClawResource) createTemplate(req *restful.Request, resp *restful.Response) {
	requester := requesterID(req)
	if requester == "" {
		resp.WriteErrorString(http.StatusUnauthorized, "unauthorized")
		return
	}
	var body openclaw.OpenClawTemplateUpsertRequest
	if err := req.ReadEntity(&body); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	data, err := (openclaw.OpenClawTemplateService{}).CreateTemplate(
		req.Request.Context(),
		requesterEmail(req),
		body,
	)
	if err != nil {
		writeOpenClawError(resp, err)
		return
	}
	common.ResponseSuccess(resp, data)
}

func (r OpenClawResource) updateTemplate(req *restful.Request, resp *restful.Response) {
	requester := requesterID(req)
	if requester == "" {
		resp.WriteErrorString(http.StatusUnauthorized, "unauthorized")
		return
	}
	var body openclaw.OpenClawTemplateUpsertRequest
	if err := req.ReadEntity(&body); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	data, err := (openclaw.OpenClawTemplateService{}).UpdateTemplate(
		req.Request.Context(),
		requesterEmail(req),
		req.PathParameter("name"),
		body,
	)
	if err != nil {
		writeOpenClawError(resp, err)
		return
	}
	common.ResponseSuccess(resp, data)
}

func buildDashboardFromInstances(instances []openclaw.OpenClawInstance, includeGatewayToken bool) openclaw.OpenClawDashboardResponse {
	resp := openclaw.OpenClawDashboardResponse{
		Overview: openclaw.OpenClawOverview{
			TotalInstances: len(instances),
		},
		Providers: make([]openclaw.OpenClawProviderBreakdown, 0),
		Instances: make([]openclaw.OpenClawInstanceTelemetry, 0, len(instances)),
	}
	providerCount := map[string]float64{}
	for i := range instances {
		item := instances[i]
		provider := providerFromInstance(item)
		if provider == "" {
			provider = "unknown"
		}
		providerCandidates := providerCandidatesFromInstance(item)
		providerBaseURL, providerAPIType, providerModelIDs := providerRuntimeDetailsFromInstance(item, provider)
		readyPods := int(item.Status.ReadyReplicas)
		if item.Accessible {
			resp.Overview.AccessibleInstances++
		}
		resp.Overview.ReadyPods += readyPods
		gatewayToken := ""
		if includeGatewayToken {
			gatewayToken = strings.TrimSpace(item.Status.GatewayToken)
		}
		resp.Instances = append(resp.Instances, openclaw.OpenClawInstanceTelemetry{
			Namespace:          item.Namespace,
			Name:               item.Name,
			DisplayName:        item.DisplayName,
			Purpose:            item.Purpose,
			OwnerID:            item.OwnerID,
			OwnerUsername:      strings.TrimSpace(item.OwnerUsername),
			OwnerEmail:         strings.TrimSpace(item.OwnerEmail),
			Accessible:         item.Accessible,
			ReadyPods:          readyPods,
			Provider:           provider,
			Endpoint:           strings.TrimSpace(item.Status.Endpoint),
			GatewayToken:       gatewayToken,
			ProviderCandidates: providerCandidates,
			ProviderPrimary:    strings.TrimSpace(item.Models.Primary),
			ProviderModelIDs:   providerModelIDs,
			ProviderBaseURL:    providerBaseURL,
			ProviderAPIType:    providerAPIType,
			UpdatedAt:          item.UpdatedAt.Format(time.RFC3339),
		})
		providerCount[provider]++
	}
	for provider, count := range providerCount {
		resp.Providers = append(resp.Providers, openclaw.OpenClawProviderBreakdown{
			Provider: provider,
			Requests: count,
		})
	}
	sort.Slice(resp.Providers, func(i, j int) bool {
		return resp.Providers[i].Requests > resp.Providers[j].Requests
	})
	return resp
}

func parseBoolQueryWithDefault(raw string, def bool) bool {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return def
	}
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}

func providerFromInstance(instance openclaw.OpenClawInstance) string {
	primary := strings.TrimSpace(instance.Models.Primary)
	if primary != "" {
		parts := strings.Split(primary, "/")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0])
		}
	}
	if len(instance.Models.Providers) > 0 {
		return strings.TrimSpace(instance.Models.Providers[0].Name)
	}
	return ""
}

func providerCandidatesFromInstance(instance openclaw.OpenClawInstance) []string {
	result := make([]string, 0, len(instance.Models.Providers))
	seen := map[string]struct{}{}
	for i := range instance.Models.Providers {
		name := strings.TrimSpace(instance.Models.Providers[i].Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func providerRuntimeDetailsFromInstance(instance openclaw.OpenClawInstance, resolvedProvider string) (string, string, []string) {
	resolvedProvider = strings.TrimSpace(resolvedProvider)
	if resolvedProvider == "" || strings.EqualFold(resolvedProvider, "unknown") {
		return "", "", nil
	}
	for i := range instance.Models.Providers {
		provider := instance.Models.Providers[i]
		name := strings.TrimSpace(provider.Name)
		if name == "" || !strings.EqualFold(name, resolvedProvider) {
			continue
		}
		modelIDs := append([]string(nil), provider.ModelIDs...)
		sort.Strings(modelIDs)
		return strings.TrimSpace(provider.BaseURL), strings.TrimSpace(provider.APIType), modelIDs
	}
	return "", "", nil
}

func requesterID(req *restful.Request) string {
	if req == nil {
		return ""
	}
	if userID := req.Attribute(config.RequestUserId); userID != nil {
		value, _ := userID.(string)
		return strings.TrimSpace(value)
	}
	return ""
}

func requesterEmail(req *restful.Request) string {
	if req == nil {
		return ""
	}
	if claimsAny := req.Attribute(config.RequestAccount); claimsAny != nil {
		if claims, ok := claimsAny.(dtos.UserClaims); ok {
			return strings.TrimSpace(claims.Email)
		}
	}
	return ""
}

func requesterOwnerMeta(req *restful.Request) openclaw.OpenClawOwnerMeta {
	meta := openclaw.OpenClawOwnerMeta{}
	if req == nil {
		return meta
	}
	if claimsAny := req.Attribute(config.RequestAccount); claimsAny != nil {
		if claims, ok := claimsAny.(dtos.UserClaims); ok {
			meta.Email = strings.TrimSpace(claims.Email)
			meta.Username = strings.TrimSpace(claims.Username)
			if meta.Username == "" {
				meta.Username = strings.TrimSpace(claims.Nickname)
			}
		}
	}
	return meta
}

func writeOpenClawError(resp *restful.Response, err error) {
	switch {
	case errors.Is(err, openclaw.ErrOpenClawForbidden):
		resp.WriteErrorString(http.StatusForbidden, "forbidden")
	case errors.Is(err, openclaw.ErrOpenClawNotFound):
		resp.WriteErrorString(http.StatusNotFound, "not found")
	case errors.Is(err, openclaw.ErrOpenClawInvalidInput):
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
	default:
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
	}
}
