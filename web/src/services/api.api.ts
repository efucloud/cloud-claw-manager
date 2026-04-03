import { request } from '@umijs/max';


//
//
//请求方法: GET
//请求地址: /
export async function getserveIndex(  options?: { [key: string]: any }) {
  return request(`/`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}
//
//
//请求方法: GET
//请求地址: /{path:*}
export async function getserveAsset(  options?: { [key: string]: any }) {
  return request(`/{path:*}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}
