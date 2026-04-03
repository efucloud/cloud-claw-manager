package config

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	"sync"
)

type tenantAuth struct {
	sync.Mutex
	verifier map[string]*oidc.IDTokenVerifier
	provider map[string]*oidc.Provider
}

func (t *tenantAuth) Add(ctx context.Context, org string, verifier *oidc.IDTokenVerifier, provider *oidc.Provider) {
	Logger.Infof("add oauth verifyer for org: %s", org)
	t.Lock()
	t.verifier[org] = verifier
	t.provider[org] = provider
	t.Unlock()
}
func (t *tenantAuth) Get(ctx context.Context, org string) (*oidc.IDTokenVerifier, *oidc.Provider) {
	return t.verifier[org], t.provider[org]
}
