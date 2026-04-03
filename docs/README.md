# OpenClaw 生产部署模板说明

本目录提供可直接用于新集群验证的生产模板文件：

1. `openclaw-template-enterprise.yaml`：OpenClaw 实例模板（含 `PVC + Service + Ingress`）。
2. `openclaw-shared-model-secret.yaml`：公用模型 API Key Secret（供所有 OpenClaw 实例共享）。

## 使用顺序

```bash
kubectl apply -f docs/openclaw-shared-model-secret.yaml
kubectl apply -f docs/openclaw-template-enterprise.yaml
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
  - `modelBaseUrl` / `modelPrimaryId`
  - `trustedProxiesJson`
  - `pvcSize` / `pvcStorageClassName`
- 为降低开源项目开箱门槛，`trustedProxiesJson` 默认值为 `["0.0.0.0/0","::/0"]`；生产环境请务必改为 ingress/LB 实际来源网段。
- 其余生产参数（探针、资源、PDB、Ingress 超时、Control UI 安全项）固定在模板正文，避免页面出现大量可调项。
- Ingress 参数由后端配置注入：
  - `ingressEnabled` / `ingressClassName` / `ingressPath` / `ingressPathType`
  - `ingressTlsEnabled` / `ingressTlsSecretName`
- OpenClaw Control UI 配对说明：
  - 中文：本仓库模板默认设置 `controlUi.dangerouslyDisableDeviceAuth=false`，建议保持为 `false`（尤其是生产环境）。
  - English: This template defaults `controlUi.dangerouslyDisableDeviceAuth=false`; keeping it `false` is recommended, especially in production.
  - 中文：当该值为 `false` 时，首次访问控制台可能出现 `pairing required`，需要管理员手动进入对应 Pod 执行 approve 后才能登录 UI。
  - English: When set to `false`, first-time UI access may show `pairing required`; an admin must exec into the target Pod and approve pairing before login.
  - 中文：仅在受控内网的临时调试场景可设置为 `true`（跳过配对/approve）。
  - English: Set it to `true` only for temporary debugging in a trusted internal environment (skip pairing/approve).
