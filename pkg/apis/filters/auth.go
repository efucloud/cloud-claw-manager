package filters

import (
	"context"
	"fmt"
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/cloud-claw-manager/pkg/dtos"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/efucloud/common"
	restful "github.com/emicklei/go-restful/v3"
)

const bearer = "bearer "

func GetRequestToken(key string, req *restful.Request) (token string) {
	authHeader := req.HeaderParameter(key)
	if len(authHeader) > 0 {
		if len(bearer) > len(authHeader) {
			return
		}
		token = authHeader[len(bearer):]
		tokenPrefix := authHeader[:len(bearer)]
		if strings.ToLower(tokenPrefix) != bearer {
			return
		}
	}
	return token
}

func GetRequestTokenFromCookie(req *restful.Request) string {
	if req == nil || req.Request == nil {
		return ""
	}
	cookie, err := req.Request.Cookie(config.SessionCookieName)
	if err != nil || cookie == nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func GetAccountClaimsFromToken(req *restful.Request) (claims dtos.UserClaims) {
	var (
		errorData common.ErrorData
	)
	ctx := context.Background()
	if req.Attribute(config.RequestContext) != nil {
		ctx = req.Attribute(config.RequestContext).(context.Context)
	}
	token := GetRequestToken(config.AuthHeader, req)
	if len(token) == 0 {
		token = req.QueryParameter("access_token")
	}
	if len(token) == 0 {
		token = GetRequestTokenFromCookie(req)
	}
	if len(token) > 0 {
		var (
			err     error
			idToken *oidc.IDToken
		)
		idToken, errorData.Err = config.SystemVerifier.Verify(ctx, token)
		if errorData.IsNotNil() {
			return
		}
		err = idToken.Claims(&claims)
		if err != nil {
			return
		}
	}

	return
}
func Auth(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	var (
		errorData common.ErrorData
	)
	lang := common.GetLanguageFromReq(req, config.RequestLanguage)
	ctx := context.Background()
	if req.Attribute(config.RequestContext) != nil {
		ctx = req.Attribute(config.RequestContext).(context.Context)
	}
	claims := GetAccountClaimsFromToken(req)
	if claims.EAuthId == "" {
		errorData.Lang = lang
		errorData.Err = fmt.Errorf("未登录或者认证信息无效")
		errorData.ResponseCode = http.StatusUnauthorized
		common.ResponseErrorMessage(ctx, req, resp, config.Bundle, errorData)
		return
	}
	ctx = context.WithValue(ctx, config.RequestUserId, claims.EAuthId)
	ctx = context.WithValue(ctx, config.RequestAccount, claims)
	req.SetAttribute(config.RequestContext, ctx)
	req.SetAttribute(config.RequestUserId, claims.EAuthId)
	req.SetAttribute(config.RequestAccount, claims)
	chain.ProcessFilter(req, resp)
}

func Permission(roles []string) func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		lang := common.GetLanguageFromReq(req, config.RequestLanguage)
		ctx := context.Background()
		var errorData common.ErrorData

		userId := req.Attribute(config.RequestUserId)
		if userId == nil {
			errorData.Lang = lang
			errorData.ResponseCode = http.StatusForbidden
			errorData.MsgCode = config.MsgCodeCurrentActionIsForbidden
			common.ResponseErrorMessage(ctx, req, resp, config.Bundle, errorData)
			return
		}
		claimsAny := req.Attribute(config.RequestAccount)
		if claimsAny == nil {
			errorData.Lang = lang
			errorData.ResponseCode = http.StatusForbidden
			errorData.MsgCode = config.MsgCodeCurrentActionIsForbidden
			common.ResponseErrorMessage(ctx, req, resp, config.Bundle, errorData)
			return
		}
		claims, ok := claimsAny.(dtos.UserClaims)
		if !ok || !common.StringKeyInArray(claims.Role, roles) {
			errorData.Lang = lang
			errorData.ResponseCode = http.StatusForbidden
			errorData.MsgCode = config.MsgCodeCurrentActionIsForbidden
			common.ResponseErrorMessage(ctx, req, resp, config.Bundle, errorData)
			return
		}
		chain.ProcessFilter(req, resp)
	}
}
