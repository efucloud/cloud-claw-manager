import { request } from '@umijs/max';


//生成数据库ID
//生成数据库ID
//请求方法: GET
//请求地址: /api/v1/generateDatabaseId
export async function generateDatabaseId(  options?: { [key: string]: any }) {
  return request(`/api/v1/generateDatabaseId`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}
//健康检查
//健康检查
//请求方法: GET
//请求地址: /api/v1/health
export async function health(  options?: { [key: string]: any }) {
  return request(`/api/v1/health`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}
//查看应用信息
//查看应用的编译信息
//请求方法: GET
//请求地址: /api/v1/info
export async function info(  options?: { [key: string]: any }) {
  return request(`/api/v1/info`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}
