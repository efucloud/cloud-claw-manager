export const DefaultJavaScript = `
// 默认参数
// Organization: 组织信息，格式为{...}，具体字段见组织详情接口
// Workspace: 作空间信息，格式为{...}，，具体字段见工作空间详情接口
// AlertData: 告警数据,格式为{...}
// 可以使用的函数
// console.log(); // 打印日志
// httpRequest(); // 发起http请求
(function(){  
  //code,返回false|0|f为false，其他为true 
  return true||false||0||'f';
})()
  `;

export const DefaultSolution = `
## 原因分析
1. 原因1
2. 原因2
## 解决方案
1. 步骤1
2. 步骤2
## 预防建议
1. 建议1
2. 建议2
`;
export const AlertHangUp = `
## 挂起原因
1. 原因1
2. 原因2
`;
export const AlertIgnore = `
## 忽略原因
1. 原因1
2. 原因2
`;
