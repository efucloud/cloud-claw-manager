package filters

import (
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/emicklei/go-restful/v3"
)

func Log(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	chain.ProcessFilter(req, resp)
	operator := req.Attribute(config.RequestUserId)
	if operator == nil {
		operator = "unknown"
	}
	config.Logger.Debugf("operator: %s method: %s request uri: %s response code: %d", operator, req.Request.Method, req.Request.URL.Path, resp.StatusCode())
}
