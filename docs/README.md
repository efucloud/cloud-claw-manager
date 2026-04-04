# OpenClaw 生产部署模板说明

本目录提供可直接用于新集群验证的生产模板文件：

1. `openclaw-template-enterprise.yaml`：OpenClaw 实例模板（含 `PVC + Service + Ingress`）。
2. `openclaw-shared-model-secret.yaml`：公用模型 API Key Secret（供所有 OpenClaw 实例共享）。
3. `openclaw-template-cloud-ide.yaml`：云端 IDE 开发模板（内置 DevOps/React/Golang 多 Agent 协作配置，含 PVC）。
4. `openclaw-template-cloud-ide-no-pvc.yaml`：云端 IDE 开发模板（无 PVC，使用 `emptyDir` 临时存储）。
5. `iframe-preview-implementation.md`：iframe 预览 + cert-manager 自动证书实施手册（含逐步检查与排障命令）。

## 使用顺序

```bash
kubectl apply -f docs/openclaw-shared-model-secret.yaml
kubectl apply -f docs/openclaw-template-enterprise.yaml
# 或：
# kubectl apply -f docs/openclaw-template-cloud-ide.yaml
# 或（无 PVC 临时存储）：
# kubectl apply -f docs/openclaw-template-cloud-ide-no-pvc.yaml
```

## 关键点

- 实例名称由后端自动生成，格式为 `openclaw-xxxxxx`。
- Gateway Token 由后端自动生成（20 位），每个实例独立，不需要用户输入。
- owner 和 instance 信息会由后端统一注入到资源标签/注解（含工作负载 PodTemplate）。
- 模型密钥通过固定公用 Secret 注入：`openclaw-shared-model-secret` / `OPENCLAW_MODEL_API_KEY`（模板内固定引用，不向企业用户暴露配置项）。
- 新建实例前请先在 `openclaw` 命名空间准备该公用 Secret；多个 OpenClaw 实例共享同一 API Key。
- `openclaw-shared-model-secret.yaml` 中默认是占位符，应用前需要替换成真实 Key。
- `AGENTS.md` 存放在 `ConfigMap`，`openclaw.json` 存放在 `Secret`，分别挂载后由 initContainer 拷贝到运行目录。
- 模板 `defaults` 已精简，仅保留少量建议给实例覆盖的字段：
  - `replicas` / `servicePort` / `image` / `gatewaySecretSuffix`
  - `modelBaseUrl` / `modelPrimaryId` / `dashboardOrigin`
  - `trustedProxiesJson`
  - `pvcSize` / `pvcStorageClassName`
- 为降低开源项目开箱门槛，`trustedProxiesJson` 默认值为 `["0.0.0.0/0","::/0"]`；生产环境请务必改为 ingress/LB 实际来源网段。
- 其余生产参数（探针、资源、PDB、Ingress 超时、Control UI 安全项）固定在模板正文，避免页面出现大量可调项。
- Ingress 参数由后端配置注入：
  - `ingressEnabled` / `ingressClassName` / `ingressPath` / `ingressPathType`
  - `ingressTlsEnabled` / `ingressTlsSecretName`
- cert-manager 证书管理（默认启用）：
  - 模板默认 `certManagerClusterIssuer=letsencrypt-dns-prod`，会自动给 Ingress 增加 `cert-manager.io/cluster-issuer` 注解。
  - 当 `certManagerClusterIssuer` 非空时，模板强制使用每实例独立证书 Secret：`<instanceName>-tls`（避免复用固定 TLS Secret 导致自动签发失效）。
  - 如需关闭自动签发，可将 `certManagerClusterIssuer` 设为空字符串。
- iframe 预览安全头：
  - 三份模板都内置了 `<instanceName>-ingress-custom-headers` ConfigMap，并通过 Ingress `nginx.ingress.kubernetes.io/custom-headers` 注入 `Content-Security-Policy`。
  - 默认策略包含 `frame-ancestors 'self' {{ .dashboardOrigin }} https://{{ .previewHost }} https://*.openclaw.efucloud.com`。
  - `dashboardOrigin` 默认值是 `https://dashboard.openclaw.efucloud.com`，如控制台域名不同请在实例 `templateValues` 覆盖。
- OpenClaw Control UI 配对说明：
  - 中文：本仓库模板默认设置 `controlUi.dangerouslyDisableDeviceAuth=false`，建议保持为 `false`（尤其是生产环境）。
  - English: This template defaults `controlUi.dangerouslyDisableDeviceAuth=false`; keeping it `false` is recommended, especially in production.
  - 中文：当该值为 `false` 时，首次访问控制台可能出现 `pairing required`，需要管理员手动进入对应 Pod 执行 approve 后才能登录 UI。
  - English: When set to `false`, first-time UI access may show `pairing required`; an admin must exec into the target Pod and approve pairing before login.
  - 中文：仅在受控内网的临时调试场景可设置为 `true`（跳过配对/approve）。
  - English: Set it to `true` only for temporary debugging in a trusted internal environment (skip pairing/approve).
- 官方参考：
  - 配置字段（`openclaw.json`）：https://docs.openclaw.ai/gateway/configuration-reference
  - Control UI 配对与鉴权：https://docs.openclaw.ai/web/control-ui

## iframe 预览实施步骤（公有云 + ingress-nginx）

1. 安装并确认 cert-manager 可用（含 DNS-01 webhook provider）。
2. 确认 `ClusterIssuer` 就绪（例如 `letsencrypt-dns-prod`）。
3. 应用 OpenClaw 模板（默认会在实例 Ingress 上启用 cert-manager 注解与独立 TLS Secret）。
4. 在 ingress-nginx 控制器 ConfigMap 打开响应头能力（避免 iframe 安全头被拦截）：

```bash
kubectl -n ingress-nginx patch cm ingress-nginx-controller --type merge -p '{
  "data": {
    "hide-headers": "X-Frame-Options,Content-Security-Policy",
    "global-allowed-response-headers": "Content-Security-Policy"
  }
}'
kubectl -n ingress-nginx rollout restart deploy/ingress-nginx-controller
kubectl -n ingress-nginx rollout status deploy/ingress-nginx-controller
```

5. 若 ingress-nginx 控制器异常重启并出现 `--profiler-port` 冲突，可禁用 profiler：

```bash
kubectl -n ingress-nginx patch deploy ingress-nginx-controller --type='json' -p='[
  {"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--profiler-port=0"}
]'
kubectl -n ingress-nginx rollout restart deploy/ingress-nginx-controller
kubectl -n ingress-nginx rollout status deploy/ingress-nginx-controller
```

6. 验证证书与 iframe 响应头：

```bash
kubectl -n openclaw get certificate
curl -kI "https://<instance-host>/" | egrep -i "x-frame-options|content-security-policy"
```

预期：`Certificate READY=True`，且不再返回 `X-Frame-Options: DENY`。
