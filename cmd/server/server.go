package server

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/efucloud/cloud-claw-manager/cmd/server/options"
	"github.com/efucloud/cloud-claw-manager/pkg/apis"
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/cloud-claw-manager/pkg/crons"
	"github.com/efucloud/cloud-claw-manager/pkg/embeds"
	"github.com/efucloud/cloud-claw-manager/pkg/kube"
	"github.com/efucloud/cloud-claw-manager/pkg/leader"
	"github.com/efucloud/common"
	"github.com/efucloud/common/signals"
	"github.com/emicklei/go-restful/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"strings"
	"time"
)

// 选举配置常量
const (
	LeaderLockName = "cloud-claw-manager-leader"
	LeaseDuration  = 15 * time.Second
	RenewDeadline  = 10 * time.Second
	RetryPeriod    = 2 * time.Second
)

func removeLastTwoSegments(s string) string {
	lastIdx := strings.LastIndex(s, "-")
	if lastIdx == -1 {
		return s
	}
	temp := s[:lastIdx]
	secondLastIdx := strings.LastIndex(temp, "-")
	if secondLastIdx == -1 {
		return temp
	}
	return temp[:secondLastIdx]
}

const serviceAccountNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

func resolveRunNamespace() string {
	return resolveRunNamespaceWithNamespaceFile(serviceAccountNamespaceFile)
}

func resolveRunNamespaceWithNamespaceFile(namespaceFile string) string {
	if ns := readNamespaceFromFile(namespaceFile); ns != "" {
		return ns
	}
	return "openclaw"
}

func readNamespaceFromFile(namespaceFile string) string {
	content, err := os.ReadFile(namespaceFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

func NewRunnerServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()
	cmd := &cobra.Command{
		Use:          "server",
		Long:         `cloud-claw-manager server`,
		Short:        "cloud-claw-manager server",
		Example:      `cloud-claw-manager server`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			return run(s, signals.SetupSignalHandler())
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&s.Config, "config", "c", "./config/config.yaml", "config file path")
	flags.StringVarP(&s.KubeConfig, "kubeconfig", "k", "", "kubeconfig file path (local debug only)")
	flags.StringVar(&s.LeaderNamespace, "leader-namespace", "", "leader election lease namespace (local debug only, in-cluster always current namespace)")
	return cmd
}

func run(o *options.ServerRunOptions, stopCh <-chan struct{}) (err error) {
	startAt := time.Now()
	if kube.RunningInCluster() {
		if strings.TrimSpace(o.KubeConfig) != "" {
			config.Logger.Warnf("running in cluster, --kubeconfig=%q is ignored, using in-cluster config", o.KubeConfig)
		}
		if strings.TrimSpace(o.LeaderNamespace) != "" {
			config.Logger.Warnf("running in cluster, --leader-namespace=%q is ignored, using current pod namespace", o.LeaderNamespace)
		}
	}
	config.RunKubeConfig = o.KubeConfig
	// === 1. 基础初始化 (所有环境都执行) ===
	common.LoadConfig(o.Config, config.ApplicationConfig)
	config.ApplicationConfig.Init()
	config.RunNamespace = resolveRunNamespace()
	config.Logger.Infof("build info GoVersion %s", config.GoVersion)
	config.Logger.Infof("build info Commit %s", config.Commit)
	config.Logger.Infof("build info BuildDate %s", config.BuildDate)
	defer func() {
		cost := time.Since(startAt)
		if err != nil {
			config.Logger.Errorf("server exiting with error, uptime=%s, err=%v", cost, err)
			return
		}
		config.Logger.Infof("server exiting normally, uptime=%s", cost)
	}()

	config.Bundle, _ = common.I18nInit(embeds.I18nFiles, config.Logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 注册 API
	apis.AddResources()
	// OIDC 初始化
	config.AuthProvider, err = oidc.NewProvider(ctx, config.ApplicationConfig.OidcConfig.Issuer)
	if err != nil {
		config.Logger.Errorf("get oidc config from: %s failed, err: %s", config.ApplicationConfig.OidcConfig.Issuer, err)
		return err
	}
	if config.AuthProvider != nil {
		oidcCfg := oidc.Config{ClientID: config.ApplicationConfig.OidcConfig.ClientId}
		config.SystemVerifier = config.AuthProvider.Verifier(&oidcCfg)
	}
	go func() {
		<-stopCh
		config.Logger.Infof("received stop signal, canceling server context")
		cancel()
	}()
	elector, err := leader.New(ctx, o.LeaderNamespace, LeaderLockName, LeaseDuration, RetryPeriod, leader.Callbacks{
		OnStartedLeading: func(leaderCtx context.Context) {
			config.Logger.Infof("acquired leadership, identity started services")
			crons.StartCronJob(leaderCtx)
			pro := prometheus.NewRegistry()
			mux := http.NewServeMux()
			mux.Handle("/metrics", promhttp.HandlerFor(pro, promhttp.HandlerOpts{}))
			mux.Handle("/", restful.DefaultContainer)
			server := &http.Server{
				Addr:    fmt.Sprintf(":%d", config.ServerPort),
				Handler: mux,
			}
			go func() {
				<-leaderCtx.Done()
				config.Logger.Infof("leader context done, shutting down http server")
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer shutdownCancel()
				if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
					config.Logger.Warnf("http server shutdown returned error: %v", shutdownErr)
				}
			}()
			config.Logger.Infof("leader http server listening on port: %d", config.ServerPort)
			if listenErr := server.ListenAndServe(); listenErr != nil && listenErr != http.ErrServerClosed {
				config.Logger.Errorf("leader http server failed: %v", listenErr)
			} else {
				config.Logger.Infof("leader http server stopped")
			}
		},
		OnStoppedLeading: func() {
			config.Logger.Warnf("lost leadership, services stopped")
		},
	})
	if err != nil {
		return err
	}
	err = elector.Run(ctx)
	if err != nil {
		config.Logger.Errorf("leader elector stopped with error: %v", err)
		return err
	}
	config.Logger.Infof("leader elector stopped")
	return nil
}
