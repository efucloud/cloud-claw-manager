# Cloud Claw Manager

[中文文档 (Chinese)](./README.zh-CN.md)

Repository: [github.com/efucloud/cloud-claw-manager](https://github.com/efucloud/cloud-claw-manager)

## Overview

Cloud Claw Manager is a Kubernetes-oriented management platform for OpenClaw instances.  
It helps teams deliver enterprise AI assistant instances in a standardized, auditable, and scalable way.

This repository uses a unified full-stack layout:

- Frontend source in `web/`
- Web assets embedded into backend (`pkg/embeds/web`)
- A single backend service serves both API and web UI

## Key Features

- OIDC authentication and identity propagation
- Template-driven instance provisioning (ConfigMap + JSON Schema + YAML templates)
- OpenClaw lifecycle management (create, list, stop/start, delete)
- Automatic ownership labels/annotations injection (owner/instance)
- User and admin dashboards
- No business database; state is aggregated from Kubernetes resources

## Repository Structure

```text
.
├── cmd/                 # Server entrypoint
├── pkg/                 # Backend code
├── web/                 # Frontend project
├── config/              # Config files
├── docs/                # Templates and docs
└── scripts/             # Build scripts
```

## Quick Start

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

Template details: [docs/README.md](docs/README.md)

4. Build and run

```bash
./scripts/build-web-embed.sh
go run ./cmd/start.go -c ./config/config.yaml
```

## GitHub Actions (Container Build Automation)

Workflow file: `.github/workflows/docker-image.yml`

- Triggers:
  - push to `main/master`
  - tag push (`v*`)
  - pull requests (build only, no push)
- Output:
  - Multi-arch image (`linux/amd64`, `linux/arm64`)
  - Published to `ghcr.io/efucloud/cloud-claw-manager`

## Contributing

Contributions are welcome via Issues and Pull Requests.  
Please include scope, rationale, and validation details in your PR.
