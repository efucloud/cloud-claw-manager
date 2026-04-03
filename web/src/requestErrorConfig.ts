import type { ResponseError } from "@/services/common";
import { deleteToken, getI18nLanguage, getToken } from "@/utils/global";
import type { RequestConfig } from "@umijs/max";
import { getLocale } from "@umijs/max";
import { notification } from "antd";

// 错误处理方案： 错误类型
enum ErrorShowType {
  SILENT = 0,
  WARN_MESSAGE = 1,
  ERROR_MESSAGE = 2,
  NOTIFICATION = 3,
  REDIRECT = 9,
}
// 与后端约定的响应数据格式
interface ResponseStructure {
  success: boolean;
  data: any;
  errorCode?: number;
  errorMessage?: string;
  showType?: ErrorShowType;
}
const authHeaderInterceptor = (url: string, options: RequestConfig) => {
  let currentLocale = getLocale();
  if (!currentLocale) {
    currentLocale = getI18nLanguage();
  }
  if (!currentLocale && typeof window !== "undefined") {
    currentLocale = localStorage.getItem("umi_locale") || "zh-CN";
  }
  if (!currentLocale) {
    currentLocale = "zh-CN";
  }
  options.headers = {
    ...options.headers,
    "X-Locale": currentLocale,
  };
  // 组织接口
  const token = getToken();
  if (token?.timestamp > Date.now()) {
    console.log("token过期");
    // 重定向到登录页
    window.location.href = `/index`;
  }
  if (token && token?.access_token) {
    // 添加全局token
    options.headers = {
      ...options.headers,
      Authorization: `Bearer ${token.access_token}`,
    };
  }
  return {
    url: `${url}`,
    options: options,
  };
};

/**
 * @name 错误处理
 * pro 自带的错误处理， 可以在这里做自己的改动
 * @doc https://umijs.org/docs/max/request#配置
 */
export const errorConfig: RequestConfig = {
  // 错误处理： umi@3 的错误处理方案。
  errorConfig: {
    // 错误抛出
    errorThrower: (res) => {
      console.log(res);
    },
    // 错误接收及处理
    errorHandler: (error: any) => {
      const { response } = error;
      if (response.request?.responseURL?.includes("EfuTest=true")) {
        // 跳过验证
      } else if (response && response.status) {
        const { data, status } = response;
        const errResponse = data as unknown as ResponseError;
        if (status === 401) {
          deleteToken();
          window.location.href = `/login`;
          return;
        } else if (status === 409) {
          if (getI18nLanguage() === "en-US") {
            notification.error({
              title: "Resource Conflict in current cluster",
              description: errResponse?.alert || errResponse?.message || "",
            });
          } else {
            notification.error({
              title: "集群中该资源冲突，可能已经存在",
              description: errResponse?.alert || errResponse?.message || "",
            });
          }
        } else if (status === 404) {
          if (response && response.data.alert) {
            notification.error({
              title: response.data.alert,
              description: errResponse?.alert || errResponse?.message || "",
            });
          } else {
            if (getI18nLanguage() === "en-US") {
              notification.error({
                title: "Resource not found in current cluster",
              });
            } else {
              notification.error({
                title: "该资源不存在",
              });
            }
          }
        } else if (status === 400) {
          if (errResponse?.alert) {
            notification.error({
              title: errResponse?.alert || errResponse?.message,
            });
          } else {
            if (getI18nLanguage() === "en-US") {
              notification.error({
                title: "Bad request, please check your request parameters",
                description: errResponse?.alert || errResponse?.message || "",
              });
            } else {
              notification.error({
                title: "错误请求，请检查请求参数",
                description: errResponse?.alert
                  ? errResponse?.alert
                  : errResponse?.message || "",
              });
            }
          }
        } else if (status === 422) {
          if (getI18nLanguage() === "en-US") {
            notification.error({
              title: "Request has Unprocessable Entity",
              description: errResponse?.message || "",
            });
          } else {
            notification.error({
              title: "请求数据中字段存在非法数据，应该为空或者不传",
              description: errResponse?.message || "",
            });
          }
        } else if (status === 403) {
          if (getI18nLanguage() === "en-US") {
            notification.error({
              title: "Request is Forbidden",
              description: errResponse?.message || "",
            });
          } else {
            notification.error({
              title: "权限不足",
              description: errResponse?.message || "",
            });
          }
        } else {
          if (
            response.request.responseURL.includes("/api/v1/kubernetes/") &&
            response.request.responseURL.includes("/proxy/")
          ) {
            if (getI18nLanguage() === "en-US") {
              notification.error({
                title: "The cluster is offline or the network is unavailable",
                description: errResponse?.message || "",
              });
            } else {
              notification.error({
                title: "集群下线或者网络不可达",
                description: errResponse?.message || "",
              });
            }
            return;
          }
          if (errResponse.alert) {
            notification.error({
              title: errResponse.alert,
              description: errResponse.detail,
            });
          } else {
            if (getI18nLanguage() === "en-US") {
              notification.error({
                title:
                  "Server connection failed, please check your network connection settings.",
              });
            } else {
              notification.error({
                title: "服务端连接失败，请检查网络连接配置",
                description: errResponse?.message || "",
              });
            }
          }
        }
      } else {
        const lang = getI18nLanguage();
        if (lang === "en-US") {
          notification.error({
            title: "Error",
            description: "Can't connect to the server",
          });
        } else {
          notification.error({
            title: "错误",
            description: "服务器连接失败",
          });
        }
      }
    },
  },

  // 请求拦截器
  requestInterceptors: [authHeaderInterceptor],

  // 响应拦截器
  responseInterceptors: [
    (response) => {
      return response;
    },
  ],
};
