package kube

import (
	"context"
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func BuildRestConfig(ctx context.Context, kubeConfig string) (*rest.Config, error) {
	if RunningInCluster() {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("build in-cluster config failed: %w", err)
		}
		return cfg, nil
	}

	if strings.TrimSpace(kubeConfig) != "" {
		if _, err := os.Stat(kubeConfig); err != nil {
			return nil, fmt.Errorf("kubeconfig file invalid: %w", err)
		}
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			return nil, fmt.Errorf("build config from kubeconfig failed: %w", err)
		}
		return cfg, nil
	}

	return nil, fmt.Errorf("kubeconfig is empty and not running in cluster")
}

func RunningInCluster() bool {
	if strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_HOST")) == "" || strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_PORT")) == "" {
		return false
	}
	paths := []string{
		"/var/run/secrets/kubernetes.io/serviceaccount/token",
		"/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		"/var/run/secrets/kubernetes.io/serviceaccount/namespace",
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}
