# Cloud Claw Manager

[English](./README.md)

仓库地址：[github.com/efucloud/cloud-claw-manager](https://github.com/efucloud/cloud-claw-manager)

## 项目简介

Cloud Claw Manager 是一个面向 Kubernetes 的 OpenClaw 实例管理平台。  
项目目标是以标准化、可审计、可扩展的方式，帮助团队交付企业级 AI 助手实例。

本仓库采用前后端合并模式：

- 前端源码位于 `web/`
- 前端构建产物嵌入后端（`pkg/embeds/web`）
- 最终以单一后端服务对外提供 API 和页面

## 核心特性

- OIDC 统一认证与身份透传
- 模板化实例创建（ConfigMap + JSON Schema + YAML 模板）
- OpenClaw 实例生命周期管理（创建、查询、启停、删除）
- 自动注入实例归属标签/注解（owner/instance）
- 用户与管理员双看板能力
- 无业务数据库，状态直接来源于 Kubernetes 资源

## 认证系统

本项目使用兼容 OIDC 的开源认证系统 `eauth`：  
[efucloud/eauth](https://github.com/efucloud/eauth)

## 目录结构

```text
.
├── cmd/                 # 启动入口
├── pkg/                 # 后端业务代码
├── web/                 # 前端工程
├── config/              # 配置文件（示例：config.yaml）
├── docs/                # 模板与部署文档
└── scripts/             # 构建脚本
```

## 快速开始

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

## GitHub Actions（自动构建镜像）

工作流文件：`.github/workflows/docker-image.yml`

- 触发条件：
  - `push` 到 `main/master`
  - `push` Tag（`v*`）
  - `pull_request`（仅构建，不推送）
- 产出：
  - `linux/amd64` + `linux/arm64` 多架构镜像
  - 推送到 `ghcr.io/efucloud/cloud-claw-manager`

## 贡献

欢迎通过 Issue / Pull Request 参与贡献。建议在 PR 中说明：

- 变更背景与目标
- 影响范围（API/模板/UI/兼容性）
- 验证方式与结果
