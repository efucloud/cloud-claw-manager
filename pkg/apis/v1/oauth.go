package v1

import (
	"context"
	"encoding/json"
	"github.com/efucloud/cloud-claw-manager/pkg/apis/filters"
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/cloud-claw-manager/pkg/dtos"
	"github.com/efucloud/cloud-claw-manager/pkg/services"
	"github.com/efucloud/cloud-claw-manager/pkg/utils"
	"net/http"
	"strings"
	"time"

	"github.com/efucloud/common"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"
)

type OAuthResource struct {
	Svc services.OAuthService
}

func (r OAuthResource) AddWebService(ws *restful.WebService) {
	apiInfo := common.ApiInfo{}
	apiInfo.Tag = "oauth"
	apiInfo.Description = "OAuth"
	apiExtend := ""
	common.RegisterApiInfo(apiInfo)
	ws.Route(ws.GET(config.APIPrefix+apiExtend+"/userinfo").
		Doc("获取用户信息").
		Notes("获取用户信息").
		Param(ws.HeaderParameter(config.AuthHeader, "请求Token").Required(true)).
		To(r.userinfo).
		Returns(http.StatusOK, "请求成功", dtos.AuthedUserInfo{}).
		Returns(http.StatusBadRequest, "请求数据无法处理", dtos.ResponseError{}).
		Returns(http.StatusInternalServerError, "内部处理逻辑错误", dtos.ResponseError{}).
		Filter(filters.I18n).Filter(filters.Log).Filter(filters.Auth).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "getUserinfo"))
	ws.Route(ws.GET(config.APIPrefix+apiExtend+"/oauth-info").
		Doc("获取认证地址").
		Notes("获取认证地址").
		To(r.oauthInfo).
		Returns(http.StatusOK, "请求成功", dtos.OidcConfig{}).
		Returns(http.StatusBadRequest, "请求数据无法处理", dtos.ResponseError{}).
		Returns(http.StatusInternalServerError, "内部处理逻辑错误", dtos.ResponseError{}).
		Filter(filters.I18n).Filter(filters.Log).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "getAuthorizeInfo"))
	ws.Route(ws.POST(config.APIPrefix+apiExtend+"/oauth/oidc").
		Doc("OIDC方式登录").
		Notes("OIDC回调后前端给到后端的Code接口，用于换取第三方的token并获取用户信息，若用户在系统不存在，"+
			"则根据组织是否允许自动注册来决定是否自动创建用户信息，若第一次是通过第三方登录，需要先设置密码，"+
			"若组织设置了MFA则需要再次输入验证码，若用户没有绑定过验证器，则返回验证器的二维码和密钥").
		To(r.loginByOIDC).
		Reads(dtos.LoginByOIDC{}).
		Returns(http.StatusOK, "成功", dtos.AccessTokenResponse{}).
		Returns(http.StatusBadRequest, "请求数据无法处理", dtos.ResponseError{}).
		Returns(http.StatusInternalServerError, "内部处理逻辑错误", dtos.ResponseError{}).
		Filter(filters.I18n).Filter(filters.Log).
		Metadata(restfulspec.KeyOpenAPITags, apiInfo.Tags()).
		Metadata(config.FrontApiTag, "loginByOidc"))
}

func (r OAuthResource) loginByOIDC(req *restful.Request, resp *restful.Response) {
	lang := common.GetLanguageFromReq(req, config.RequestLanguage)
	var (
		errorData  common.ErrorData
		result     dtos.AccessTokenResponse
		loginParam dtos.LoginByOIDC
	)
	ctx := context.Background()
	if reqCtx := req.Attribute(config.RequestContext); reqCtx != nil {
		ctx = reqCtx.(context.Context)
	}
	ctx = context.WithValue(ctx, config.RequestLanguage, lang)
	errorData.Err = json.NewDecoder(req.Request.Body).Decode(&loginParam)
	if errorData.IsNotNil() {
		config.Logger.Error(errorData.Err)
		errorData.MsgCode = config.MsgCodeJsonDecodeFailed
		errorData.ResponseCode = http.StatusBadRequest
		errorData.Lang = lang
		common.ResponseErrorMessage(ctx, req, resp, config.Bundle, errorData)
		return
	}
	result, errorData = r.Svc.LoginByOIDC(ctx, loginParam)
	if errorData.IsNotNil() {
		config.Logger.Error(errorData.Err)
		errorData.Lang = lang
		common.ResponseErrorMessage(ctx, req, resp, config.Bundle, errorData)
		return
	}
	tokenValue := result.IDToken
	if tokenValue == "" {
		tokenValue = result.AccessToken
	}
	if tokenValue != "" {
		cookie := &http.Cookie{
			Name:     config.SessionCookieName,
			Value:    tokenValue,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		if result.ExpiresIn > 0 {
			expireAt := time.Unix(result.ExpiresIn, 0)
			cookie.Expires = expireAt
			maxAge := int(time.Until(expireAt).Seconds())
			if maxAge > 0 {
				cookie.MaxAge = maxAge
			}
		}
		if req.Request != nil && (req.Request.TLS != nil || strings.EqualFold(req.Request.Header.Get("X-Forwarded-Proto"), "https")) {
			cookie.Secure = true
		}
		http.SetCookie(resp.ResponseWriter, cookie)
	}
	common.ResponseSuccess(resp, result)
}
func (r OAuthResource) oauthInfo(req *restful.Request, resp *restful.Response) {
	common.ResponseSuccess(resp, dtos.OidcConfig{
		Issuer:   config.ApplicationConfig.OidcConfig.Issuer,
		ClientId: config.ApplicationConfig.OidcConfig.ClientId,
	})
}

func (r OAuthResource) userinfo(req *restful.Request, resp *restful.Response) {
	lang := common.GetLanguageFromReq(req, config.RequestLanguage)
	var (
		errorData common.ErrorData
		userinfo  dtos.AuthedUserInfo
	)
	ctx := context.Background()
	if reqCtx := req.Attribute(config.RequestContext); reqCtx != nil {
		ctx = reqCtx.(context.Context)
	}
	ctx = context.WithValue(ctx, config.RequestLanguage, lang)
	userId := req.Attribute(config.RequestUserId)
	if userId == nil {
		errorData.MsgCode = config.MsgCodeUserInfoIsEmpty
		errorData.Lang = lang
		common.ResponseErrorMessage(ctx, req, resp, config.Bundle, errorData)
		return
	}
	userinfo, errorData = r.Svc.Userinfo(ctx, userId.(string))
	if errorData.IsNotNil() {
		config.Logger.Error(errorData.Err)
		errorData.Lang = lang
		common.ResponseErrorMessage(ctx, req, resp, config.Bundle, errorData)
		return
	}
	userinfo.RemoteAddress = utils.GetRemoteAddress(req.Request.RemoteAddr)
	common.ResponseSuccess(resp, userinfo)
}
