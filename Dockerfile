# 构建阶段
FROM golang:1.25.5 AS builder

# buildx 自动传入的跨平台参数（无需手动传）
ARG TARGETOS
ARG TARGETARCH


ARG GIT_COMMIT
ARG BUILD_DATE

WORKDIR /workspace
COPY . .

ENV GOPROXY='https://goproxy.cn,direct'
ENV GOSUMDB='off'

# 动态设置 GOOS/GOARCH（注意：Go 不支持所有 TARGETARCH 名称，需映射）
RUN case "$TARGETARCH" in \
        "amd64") GOARCH="amd64" ;; \
        "arm64") GOARCH="arm64" ;; \
        "arm")   GOARCH="arm" ;; \
        *) echo "Unsupported TARGETARCH: $TARGETARCH"; exit 1 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$GOARCH go build \
      -ldflags "-X 'github.com/efucloud/cloud-claw-manager/pkg/config.GoVersion=1.25' \
                -X 'github.com/efucloud/cloud-claw-manager/pkg/config.Commit=${GIT_COMMIT}' \
                -X 'github.com/efucloud/cloud-claw-manager/pkg/config.BuildDate=${BUILD_DATE}' \
                -X 'github.com/efucloud/cloud-claw-manager/pkg/config.Edition=community'" \
      -o ./output/cloud-claw-manager-$TARGETOS-$TARGETARCH \
      ./cmd/start.go


# 运行阶段
FROM alpine:3.23.0

# 重新声明 ARG（每个 FROM 都是新作用域）
ARG TARGETOS
ARG TARGETARCH
ARG GIT_COMMIT
ARG BUILD_DATE

# 设置 OCI 标准 Labels
LABEL org.opencontainers.image.source=https://github.com/efucloud/cloud-claw-manager
LABEL org.opencontainers.image.revision=${GIT_COMMIT}
LABEL org.opencontainers.image.created=${BUILD_DATE}
LABEL com.efucloud.build.commit=${GIT_COMMIT}
LABEL com.efucloud.build.date=${BUILD_DATE}

WORKDIR /efucloud

# 复制对应平台的二进制
COPY --from=builder /workspace/output/cloud-claw-manager-$TARGETOS-$TARGETARCH /usr/local/bin/cloud-claw-manager


EXPOSE 9004

ENTRYPOINT ["/usr/local/bin/cloud-claw-manager", "-c", "./config/config.yaml"]
