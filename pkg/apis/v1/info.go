package v1

import (
	"github.com/efucloud/cloud-claw-manager/pkg/apis/filters"
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/cloud-claw-manager/pkg/dtos"
	"github.com/efucloud/cloud-claw-manager/pkg/utils"
	"github.com/efucloud/common"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"net/http"
)

type InfoResource struct {
}

func (r InfoResource) AddWebService(ws *restful.WebService) {
	apiInfo := common.ApiInfo{}
	apiInfo.Tag = "system-info"
	apiInfo.Description = "应用信息"
	common.RegisterApiInfo(apiInfo)
	ws.Route(ws.GET(config.APIPrefix+"/health").
		Doc("健康检查").
		Notes("健康检查").
		To(r.health).
		Returns(http.StatusOK, "成功", "ok").
		Returns(http.StatusInternalServerError, "内部处理逻辑错误", common.ResponseError{}).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "health"))
	ws.Route(ws.GET(config.APIPrefix+"/info").
		Doc("查看应用信息").
		Notes("查看应用的编译信息").
		To(r.info).
		Returns(http.StatusOK, "成功", dtos.ApplicationInfo{}).
		Returns(http.StatusInternalServerError, "内部处理逻辑错误", common.ResponseError{}).
		Filter(filters.I18n).Filter(filters.Log).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "info"))

	ws.Route(ws.GET(config.APIPrefix+"/generateDatabaseId").
		Doc("生成数据库ID").
		Notes(`生成数据库ID`).
		To(r.generateDatabaseId).
		Returns(http.StatusOK, "ok", "").
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "generateDatabaseId"))

}
func (r InfoResource) generateDatabaseId(req *restful.Request, resp *restful.Response) {
	common.ResponseSuccess(resp, utils.GenerateDatabaseId())
}
func (r InfoResource) info(req *restful.Request, resp *restful.Response) {
	var app dtos.ApplicationInfo
	app.Application = config.ApplicationName
	app.BuildDate = config.BuildDate
	app.GoVersion = config.GoVersion
	app.Commit = config.Commit
	common.ResponseSuccess(resp, app)
}

func (r InfoResource) health(req *restful.Request, resp *restful.Response) {
	common.ResponseSuccess(resp, "ok")
}
