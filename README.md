# Cloud Claw Manager

[中文](#中文) | [English](#english)

Repository: [github.com/efucloud/cloud-claw-manager](https://github.com/efucloud/cloud-claw-manager)

---

## 中文

### 项目简介

Cloud Claw Manager 是一个面向 Kubernetes 的 OpenClaw 实例管理平台。  
项目目标是以标准化、可审计、可扩展的方式，帮助团队交付企业级 AI 助手实例。

本仓库采用前后端合并模式：

- 前端源码位于 `web/`
- 前端构建产物嵌入后端（`pkg/embeds/web`）
- 最终以单一后端服务对外提供 API 和页面

### 核心特性

- OIDC 统一认证与身份透传
- 模板化实例创建（ConfigMap + JSON Schema + YAML 模板）
- OpenClaw 实例生命周期管理（创建、查询、启停、删除）
- 自动注入实例归属标签/注解（owner/instance）
- 用户与管理员双看板能力
- 无业务数据库，状态直接来源于 Kubernetes 资源

### 目录结构

```text
.
├── cmd/                 # 启动入口
├── pkg/                 # 后端业务代码
├── web/                 # 前端工程
├── config/              # 配置文件（示例：config.yaml）
├── docs/                # 模板与部署文档
└── scripts/             # 构建脚本
```

### 快速开始

1. 环境准备

- Go（版本与 `go.mod` 保持一致）
- Node.js 20+
- Yarn 1.x
- Kubernetes 集群访问权限
- OIDC 提供方

2. 编辑配置 `config/config.yaml`

- `logConfig`：日志配置（已支持默认值，通常只需设置 `level`）
- `oidcConfig`：OIDC 客户端配置
- `openClawControl`：预览域名、管理员邮箱、Ingress/TLS 控制项

3. 安装推荐模板

```bash
kubectl apply -f docs/openclaw-shared-model-secret.yaml
kubectl apply -f docs/openclaw-template-enterprise.yaml
```

模板说明见 [docs/README.md](docs/README.md)。

4. 构建并启动

```bash
./scripts/build-web-embed.sh
go run ./cmd/start.go -c ./config/config.yaml
```

### GitHub Actions（自动构建镜像）

仓库已新增工作流：`.github/workflows/docker-image.yml`

- 触发条件：
  - `push` 到 `main/master`
  - `push` Tag（`v*`）
  - `pull_request`（仅构建，不推送）
- 构建内容：
  - 自动构建前端并嵌入后端
  - `linux/amd64` + `linux/arm64` 多架构镜像
- 推送地址：
  - `ghcr.io/efucloud/cloud-claw-manager`

> 注意：推送镜像需要仓库具备 `packages: write` 权限（工作流已声明该权限）。

### 贡献

欢迎通过 Issue / Pull Request 参与贡献。建议在 PR 中说明：

- 变更背景与目标
- 影响范围（API/模板/UI/兼容性）
- 验证方式与结果

---

## English

### Overview

Cloud Claw Manager is a Kubernetes-oriented management platform for OpenClaw instances.  
It helps teams deliver enterprise AI assistant instances in a standardized, auditable, and scalable way.

This repository uses a unified full-stack layout:

- Frontend source in `web/`
- Web assets embedded into backend (`pkg/embeds/web`)
- A single backend service serves both API and web UI

### Key Features

- OIDC authentication and identity propagation
- Template-driven instance provisioning (ConfigMap + JSON Schema + YAML templates)
- OpenClaw lifecycle management (create, list, stop/start, delete)
- Automatic ownership labels/annotations injection (owner/instance)
- User and admin dashboards
- No business database; state is aggregated from Kubernetes resources

### Quick Start

1. Prerequisites

- Go (compatible with `go.mod`)
- Node.js 20+
- Yarn 1.x
- Kubernetes cluster access
- OIDC provider

2. Configure `config/config.yaml`

- `logConfig`: logging settings (defaults are applied; usually `level` is enough)
- `oidcConfig`: OIDC client settings
- `openClawControl`: preview domain, admin emails, ingress/tls controls

3. Apply recommended templates

```bash
kubectl apply -f docs/openclaw-shared-model-secret.yaml
kubectl apply -f docs/openclaw-template-enterprise.yaml
```

See [docs/README.md](docs/README.md) for template details.

4. Build and run

```bash
./scripts/build-web-embed.sh
go run ./cmd/start.go -c ./config/config.yaml
```

### GitHub Actions (Container Build Automation)

Workflow file: `.github/workflows/docker-image.yml`

- Triggers:
  - push to `main/master`
  - tag push (`v*`)
  - pull requests (build only, no push)
- Output:
  - Multi-arch image (`linux/amd64`, `linux/arm64`)
  - Published to `ghcr.io/efucloud/cloud-claw-manager`

### Contributing

Contributions are welcome via Issues and Pull Requests.  
Please include scope, rationale, and validation details in your PR.
