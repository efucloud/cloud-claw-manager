package services

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/cloud-claw-manager/pkg/dtos"
	"github.com/efucloud/common"
	"golang.org/x/oauth2"
)

type OAuthService struct {
}

func (svc *OAuthService) Userinfo(ctx context.Context, userId string) (userinfo dtos.AuthedUserInfo, errorData common.ErrorData) {
	userinfo.ID = userId
	if claimsAny := ctx.Value(config.RequestAccount); claimsAny != nil {
		if claims, ok := claimsAny.(dtos.UserClaims); ok {
			userinfo.ID = claims.EAuthId
			userinfo.Username = claims.Username
			userinfo.Nickname = claims.Nickname
			userinfo.Role = claims.Role
			userinfo.Email = claims.Email
			userinfo.Phone = claims.Phone
		}
	}
	userinfo.Enable = true
	return
}

func (svc *OAuthService) LoginByOIDC(ctx context.Context, loginParam dtos.LoginByOIDC) (response dtos.AccessTokenResponse, errorData common.ErrorData) {
	var (
		token      *oauth2.Token
		userInfo   *oidc.UserInfo
		systemUser dtos.AuthedUserInfo
	)
	config.Logger.Infof("LoginByOIDC code: %s", loginParam.Code)
	oauthCfg := &oauth2.Config{
		ClientID:     config.ApplicationConfig.OidcConfig.ClientId,
		ClientSecret: config.ApplicationConfig.OidcConfig.ClientSecret,
		Endpoint:     config.AuthProvider.Endpoint(),
		RedirectURL:  loginParam.RedirectUri,
	}
	token, errorData.Err = oauthCfg.Exchange(ctx, loginParam.Code)
	if errorData.IsNotNil() {
		config.Logger.Error(errorData.Err)
		return
	} else {
		userInfo, errorData.Err = config.AuthProvider.UserInfo(ctx, oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: token.AccessToken,
			TokenType:   "Bearer", // The UserInfo endpoint requires a bearer token as per RFC6750
		}))
		errorData.Err = userInfo.Claims(&systemUser)
		if errorData.IsNotNil() {
			config.Logger.Error(errorData.Err)
			return
		}
	}

	response.AccessToken = token.AccessToken
	if idToken, ok := token.Extra("id_token").(string); ok {
		response.IDToken = idToken
	}
	response.TokenType = token.TokenType
	response.RefreshToken = token.RefreshToken
	response.ExpiresIn = token.Expiry.Unix()
	return
}
