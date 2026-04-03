package apis

import (
	"fmt"
	v1 "github.com/efucloud/cloud-claw-manager/pkg/apis/v1"
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/common"
	"github.com/emicklei/go-restful/v3"
	"net/http"
)

func GetWebServices(container *restful.Container) *restful.WebService {
	ws := new(restful.WebService)
	container.RecoverHandler(func(i interface{}, writer http.ResponseWriter) {
		writer.WriteHeader(http.StatusInternalServerError)
		var body common.ResponseError
		body.Detail = fmt.Sprintf("%v", i)
		body.Alert, _ = common.GetLocaleMessage(config.Bundle, nil, "zh", config.MsgCodeApplicationUnExceptPanicError)
	})
	ws.Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	v1.InfoResource{}.AddWebService(ws)
	v1.OAuthResource{}.AddWebService(ws)
	v1.OpenClawResource{}.AddWebService(ws)
	StaticResource{}.AddWebService(ws)

	return ws
}
func AddResources() {
	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.DefaultResponseContentType(restful.MIME_JSON)
	container := restful.DefaultContainer

	container.Router(restful.CurlyRouter{})
	container.Filter(container.OPTIONSFilter)
	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept", "*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "PATCH", "*"},
		CookiesAllowed: true,
		Container:      container,
	}
	container.Filter(cors.Filter)
	ws := GetWebServices(container)
	container.Add(ws)

}
