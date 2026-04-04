# OpenClaw iframe 预览实施手册（公有云 + ingress-nginx + cert-manager）

本文用于在公有云 Kubernetes 集群中，稳定实现：

- `https://dashboard.openclaw.efucloud.com/dashboard` 内嵌 iframe 预览实例
- 实例域名自动签发 TLS 证书（cert-manager）

## 0. 前置检查

### 执行

```bash
kubectl config current-context
kubectl get ns
kubectl get deploy -A | grep -E "ingress-nginx-controller|cert-manager"
```

### 检查

- 当前上下文是公有云生产/测试集群（不是本地 `kind-kind`）。
- `ingress-nginx` 和 `cert-manager` 相关 Deployment 存在。

### 通过标准

- 能看到 `ingress-nginx-controller` 和 `cert-manager` 组件。

---

## 1. cert-manager 与 DNS-01 通道可用

### 执行

```bash
kubectl get apiservice | grep dns.aliyun.com
kubectl describe clusterissuer letsencrypt-dns-prod
kubectl -n cert-manager get deploy,po,svc,ep | grep -i alidns
```

### 检查

- `v1alpha1.dns.aliyun.com` 的 `AVAILABLE` 为 `True`。
- `ClusterIssuer/letsencrypt-dns-prod` 处于 `Ready=True`。
- `cert-manager-webhook-alidns` 有 Running Pod，Service 有 Endpoints。

### 通过标准

- DNS-01 webhook 可调用，不再出现 `MissingEndpoints`。

---

## 2. ingress-nginx 开启 iframe 所需响应头能力

> 这一步是为了确保 Ingress 可以移除上游 `X-Frame-Options`，并允许注入自定义 CSP。

### 执行

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

### 检查

```bash
kubectl -n ingress-nginx get cm ingress-nginx-controller -o jsonpath='{.data.hide-headers}{"\n"}{.data.global-allowed-response-headers}{"\n"}'
kubectl -n ingress-nginx get pods
```

### 通过标准

- `hide-headers` 包含 `X-Frame-Options,Content-Security-Policy`
- `global-allowed-response-headers` 包含 `Content-Security-Policy`
- ingress-nginx controller Pod 全部 `Running/Ready`

### 常见异常

- 若 controller 日志出现 `port 10245 is already in use`：

```bash
kubectl -n ingress-nginx patch deploy ingress-nginx-controller --type='json' -p='[
  {"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--profiler-port=0"}
]'
kubectl -n ingress-nginx rollout restart deploy/ingress-nginx-controller
kubectl -n ingress-nginx rollout status deploy/ingress-nginx-controller
```

---

## 3. 应用 OpenClaw 模板

模板已内置以下能力：

- cert-manager 注解：`cert-manager.io/cluster-issuer`
- 每实例独立 TLS Secret：`<instanceName>-tls`
- iframe CSP 注入（`custom-headers` + `ConfigMap`）

### 执行

```bash
kubectl apply -f docs/openclaw-shared-model-secret.yaml
kubectl apply -f docs/openclaw-template-enterprise.yaml
# 或：
# kubectl apply -f docs/openclaw-template-cloud-ide.yaml
# 或：
# kubectl apply -f docs/openclaw-template-cloud-ide-no-pvc.yaml
```

### 检查

```bash
kubectl -n openclaw get configmap | grep openclaw-template
```

### 通过标准

- 模板 ConfigMap 成功更新。

---

## 4. 创建实例后检查 Ingress 与证书自动签发

### 执行

```bash
kubectl -n openclaw get ing
kubectl -n openclaw get certificate
```

### 检查（以实例 `openclaw-xxxxxx` 为例）

```bash
kubectl -n openclaw get ing openclaw-xxxxxx -o jsonpath='{.metadata.annotations.cert-manager\.io/cluster-issuer}{"\n"}{.spec.tls[0].secretName}{"\n"}{.metadata.annotations.nginx\.ingress\.kubernetes\.io/proxy-hide-headers}{"\n"}{.metadata.annotations.nginx\.ingress\.kubernetes\.io/custom-headers}{"\n"}'
kubectl -n openclaw get certificate openclaw-xxxxxx-tls
kubectl -n openclaw get secret openclaw-xxxxxx-tls
```

### 通过标准

- Ingress 注解中有 `cert-manager.io/cluster-issuer`
- TLS Secret 为 `openclaw-xxxxxx-tls`
- `Certificate` 为 `READY=True`
- 对应 TLS Secret 已创建

---

## 5. 检查 iframe 响应头是否正确

### 执行

```bash
curl -kI "https://openclaw-xxxxxx-openclaw.openclaw.efucloud.com/" | egrep -i "x-frame-options|content-security-policy"
```

### 检查

- 不应再出现 `X-Frame-Options: DENY`
- `Content-Security-Policy` 应包含：
  - `frame-ancestors 'self'`
  - `https://dashboard.openclaw.efucloud.com`（或你实际 dashboard 域名）

### 通过标准

- 在 dashboard 页面 iframe 内可以正常渲染实例页面。

---

## 6. 故障定位速查

### 现象 A：证书一直不签发（`Certificate READY=False`）

```bash
kubectl -n openclaw describe certificate <name>
kubectl -n openclaw get certificaterequest,order,challenge
kubectl -n openclaw describe challenge
```

重点看：

- `Presented: false` 且报 `post alidns-solver.dns.aliyun.com`
- APIService `v1alpha1.dns.aliyun.com` 是否 `Available=True`

---

### 现象 B：浏览器可直接打开，iframe 白屏/拒绝嵌入

```bash
kubectl -n openclaw get ing <instance> -o yaml
kubectl -n openclaw get cm <instance>-ingress-custom-headers -o yaml
curl -kI "https://<instance-host>/" | egrep -i "x-frame-options|content-security-policy"
```

重点看：

- Ingress 是否有 `proxy-hide-headers` 与 `custom-headers`
- 返回头是否仍有 `X-Frame-Options: DENY`

---

### 现象 C：503 Service Temporarily Unavailable

```bash
kubectl -n openclaw get ing <instance> -o yaml
kubectl -n openclaw get svc <instance> -o yaml
kubectl -n openclaw get ep <instance> -o yaml
kubectl -n openclaw get pod -l app.kubernetes.io/instance=<instance> -owide
kubectl -n ingress-nginx logs deploy/ingress-nginx-controller --since=15m | grep -E "<instance>|503|upstream|no endpoints"
```

重点看：

- Endpoints 是否为空
- ingress-nginx 是否报 upstream/no endpoints 错误

---

## 7. 建议的实例模板覆盖值（可选）

创建实例时建议显式传入：

```json
{
  "certManagerClusterIssuer": "letsencrypt-dns-prod",
  "dashboardOrigin": "https://dashboard.openclaw.efucloud.com"
}
```

当 dashboard 域名变化时，仅需修改 `dashboardOrigin`。
