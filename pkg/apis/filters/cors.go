package filters

import (
	"github.com/emicklei/go-restful/v3"
	"net/http"
)

func Cors(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	resp.AddHeader("Access-Control-Allow-Origin", req.Request.Header.Get("Origin"))
	resp.AddHeader("Access-Control-Allow-Credentials", "true")
	resp.AddHeader("Vary", "Origin")
	// 处理预检请求
	if req.Request.Method == "OPTIONS" {
		resp.AddHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		resp.AddHeader("Access-Control-Allow-Headers", "Authorization, Content-Type")
		resp.WriteHeader(http.StatusOK)
		return
	}
	chain.ProcessFilter(req, resp)
}
