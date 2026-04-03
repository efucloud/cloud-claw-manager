package apis

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/efucloud/cloud-claw-manager/pkg/embeds"
	"github.com/emicklei/go-restful/v3"
)

type StaticResource struct{}

func (r StaticResource) AddWebService(ws *restful.WebService) {
	ws.Route(ws.GET("/").To(r.serveIndex))
	ws.Route(ws.GET("/{path:*}").To(r.serveAsset))
}

func (r StaticResource) serveIndex(req *restful.Request, resp *restful.Response) {
	r.serveFile("index.html", resp)
}

func (r StaticResource) serveAsset(req *restful.Request, resp *restful.Response) {
	assetPath := strings.TrimSpace(req.PathParameter("path"))
	// /api 前缀只保留给后端接口
	if strings.HasPrefix(assetPath, "api/") || assetPath == "api" {
		resp.WriteErrorString(http.StatusNotFound, "not found")
		return
	}
	assetPath = path.Clean("/" + assetPath)
	assetPath = strings.TrimPrefix(assetPath, "/")
	if assetPath == "" || assetPath == "." {
		assetPath = "index.html"
	}
	if assetPath == "workplace" || strings.HasPrefix(assetPath, "workplace/") || assetPath == "workspace" || strings.HasPrefix(assetPath, "workspace/") {
		resp.WriteErrorString(http.StatusNotFound, "not found")
		return
	}

	if err := r.serveFile(assetPath, resp); err != nil {
		// SPA 路由回退（无扩展名时返回 index.html）
		if !strings.Contains(path.Base(assetPath), ".") {
			r.serveFile("index.html", resp)
			return
		}
		resp.WriteErrorString(http.StatusNotFound, "not found")
	}
}

func (r StaticResource) serveFile(assetPath string, resp *restful.Response) error {
	content, err := fs.ReadFile(embeds.WebFiles, "web/"+assetPath)
	if err != nil {
		return err
	}
	if ct := mime.TypeByExtension(path.Ext(assetPath)); ct != "" {
		resp.AddHeader("Content-Type", ct)
	}
	_, err = resp.Write(content)
	return err
}
