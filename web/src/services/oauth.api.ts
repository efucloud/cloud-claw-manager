import { request } from '@umijs/max';


//获取认证地址
//获取认证地址
//请求方法: GET
//请求地址: /api/v1/oauth-info
export async function getAuthorizeInfo(  options?: { [key: string]: any }) {
  return request(`/api/v1/oauth-info`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}
//获取用户信息
//获取用户信息
//请求方法: GET
//请求地址: /api/v1/userinfo
export async function getUserinfo<AuthedUserInfo>(  options?: { [key: string]: any }) {
  return request<AuthedUserInfo>(`/api/v1/userinfo`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}
//OIDC方式登录
//OIDC回调后前端给到后端的Code接口，用于换取第三方的token并获取用户信息，若用户在系统不存在，则根据组织是否允许自动注册来决定是否自动创建用户信息，若第一次是通过第三方登录，需要先设置密码，若组织设置了MFA则需要再次输入验证码，若用户没有绑定过验证器，则返回验证器的二维码和密钥
//请求方法: POST
//请求地址: /api/v1/oauth/oidc
export async function loginByOidc<AccessTokenResponse>(  data: LoginByOIDC,   options?: { [key: string]: any }) {
  return request<AccessTokenResponse>(`/api/v1/oauth/oidc`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}
