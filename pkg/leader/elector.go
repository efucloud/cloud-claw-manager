package leader

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/cloud-claw-manager/pkg/kube"
	coordinationv1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Callbacks struct {
	OnStartedLeading func(ctx context.Context)
	OnStoppedLeading func()
}

type Elector struct {
	leaseName      string
	leaseNamespace string
	identity       string
	leaseDuration  time.Duration
	retryPeriod    time.Duration
	client         *kubernetes.Clientset
	callbacks      Callbacks
}

func New(ctx context.Context, preferredNamespace, leaseName string, leaseDuration, retryPeriod time.Duration, callbacks Callbacks) (*Elector, error) {
	restCfg, err := kube.BuildRestConfig(ctx, config.RunKubeConfig)
	if err != nil {
		return nil, fmt.Errorf("build rest config for leader elector failed: %w", err)
	}
	client, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client failed: %w", err)
	}
	namespace := resolveLeaseNamespace(preferredNamespace)
	identity := detectIdentity()
	return &Elector{
		leaseName:      leaseName,
		leaseNamespace: namespace,
		identity:       identity,
		leaseDuration:  leaseDuration,
		retryPeriod:    retryPeriod,
		client:         client,
		callbacks:      callbacks,
	}, nil
}

func (e *Elector) Run(ctx context.Context) error {
	var (
		isLeader   bool
		leaderCtx  context.Context
		leaderStop context.CancelFunc
	)
	tryElect := func() {
		ok, err := e.acquireOrRenew(ctx)
		if err != nil {
			config.Logger.Warnf("leader election retry failed, lease=%s namespace=%s err=%v", e.leaseName, e.leaseNamespace, err)
			return
		}
		if ok && !isLeader {
			isLeader = true
			leaderCtx, leaderStop = context.WithCancel(ctx)
			config.Logger.Infof("leader election acquired, lease=%s namespace=%s identity=%s", e.leaseName, e.leaseNamespace, e.identity)
			if e.callbacks.OnStartedLeading != nil {
				go e.callbacks.OnStartedLeading(leaderCtx)
			}
			return
		}
		if !ok && isLeader {
			isLeader = false
			leaderStop()
			config.Logger.Warnf("leader election lost, lease=%s namespace=%s identity=%s", e.leaseName, e.leaseNamespace, e.identity)
			if e.callbacks.OnStoppedLeading != nil {
				e.callbacks.OnStoppedLeading()
			}
		}
	}

	tryElect()
	ticker := time.NewTicker(e.retryPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			if isLeader {
				leaderStop()
				if e.callbacks.OnStoppedLeading != nil {
					e.callbacks.OnStoppedLeading()
				}
			}
			return nil
		case <-ticker.C:
			tryElect()
		}
	}
}

func (e *Elector) acquireOrRenew(ctx context.Context) (bool, error) {
	lease, exists, err := e.getLease(ctx)
	if err != nil {
		return false, err
	}
	now := time.Now().UTC()
	if !exists {
		if err = e.createLease(ctx, now); err != nil {
			if apierrors.IsAlreadyExists(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}

	holder := strings.TrimSpace(derefString(lease.Spec.HolderIdentity))
	leaseSeconds := int32(e.leaseDuration.Seconds())
	if lease.Spec.LeaseDurationSeconds != nil && *lease.Spec.LeaseDurationSeconds > 0 {
		leaseSeconds = *lease.Spec.LeaseDurationSeconds
	}
	renewTime := leaseTime(lease.Spec.RenewTime)
	if renewTime.IsZero() {
		renewTime = leaseTime(lease.Spec.AcquireTime)
	}
	expired := renewTime.IsZero() || now.After(renewTime.Add(time.Duration(leaseSeconds)*time.Second))

	if holder == e.identity || expired {
		if err = e.updateLease(ctx, lease, now); err != nil {
			if apierrors.IsConflict(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (e *Elector) getLease(ctx context.Context) (*coordinationv1.Lease, bool, error) {
	lease, err := e.client.CoordinationV1().Leases(e.leaseNamespace).Get(ctx, e.leaseName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return lease, true, nil
}

func (e *Elector) createLease(ctx context.Context, now time.Time) error {
	seconds := int32(e.leaseDuration.Seconds())
	t := metav1.NewMicroTime(now)
	_, err := e.client.CoordinationV1().Leases(e.leaseNamespace).Create(ctx, &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      e.leaseName,
			Namespace: e.leaseNamespace,
		},
		Spec: coordinationv1.LeaseSpec{
			HolderIdentity:       stringPtr(e.identity),
			LeaseDurationSeconds: &seconds,
			AcquireTime:          &t,
			RenewTime:            &t,
		},
	}, metav1.CreateOptions{})
	return err
}

func (e *Elector) updateLease(ctx context.Context, current *coordinationv1.Lease, now time.Time) error {
	next := current.DeepCopy()
	seconds := int32(e.leaseDuration.Seconds())
	t := metav1.NewMicroTime(now)
	next.Spec.HolderIdentity = stringPtr(e.identity)
	next.Spec.LeaseDurationSeconds = &seconds
	if next.Spec.AcquireTime == nil {
		next.Spec.AcquireTime = &t
	}
	next.Spec.RenewTime = &t
	_, err := e.client.CoordinationV1().Leases(e.leaseNamespace).Update(ctx, next, metav1.UpdateOptions{})
	return err
}

func resolveLeaseNamespace(preferred string) string {
	return resolveLeaseNamespaceWithNamespaceFile(preferred, "/var/run/secrets/kubernetes.io/serviceaccount/namespace")
}

func resolveLeaseNamespaceWithNamespaceFile(preferred, namespaceFile string) string {
	_ = preferred
	if ns := strings.TrimSpace(config.RunNamespace); ns != "" {
		return ns
	}
	if ns := readNamespaceFromFile(namespaceFile); ns != "" {
		config.RunNamespace = ns
		return ns
	}
	return "openclaw"
}

func readNamespaceFromFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	ns := strings.TrimSpace(string(b))
	return ns
}

func detectIdentity() string {
	hostname, _ := os.Hostname()
	if strings.TrimSpace(hostname) == "" {
		hostname = "cloud-claw-manager"
	}
	return fmt.Sprintf("%s-%d", hostname, os.Getpid())
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringPtr(s string) *string {
	return &s
}

func leaseTime(t *metav1.MicroTime) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.Time
}
