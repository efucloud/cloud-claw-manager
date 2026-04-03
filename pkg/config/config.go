/*
Copyright 2022 The itcloudy.com Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	ulid "github.com/oklog/ulid/v2"
	"go.uber.org/zap"
	"math/rand"
	"strings"
	"sync"
	"time"
)

var (
	GoVersion string
	Commit    string
	BuildDate string
	Edition   string
)
var (
	ChatBuiltin          string
	ApplicationConfig    *Config
	configOnce           sync.Once
	AuthProvider         *oidc.Provider
	SystemVerifier       *oidc.IDTokenVerifier
	Bundle               *i18n.Bundle
	TenantAuth           *tenantAuth
	Logger               *zap.SugaredLogger
	RunNamespace         string
	RunKubeConfig        string
	ServerName           string
	KubeSystemCreateTime string
	KubeSystemUID        string
)

func generateDatabaseId() string {
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	id, _ := ulid.New(ms, entropy)
	m := strings.ToLower(id.String())
	return m
}
func init() {
	ChatBuiltin = generateDatabaseId()
	if ApplicationConfig == nil {
		configOnce.Do(func() {
			ApplicationConfig = new(Config)
			TenantAuth = new(tenantAuth)
			TenantAuth.provider = make(map[string]*oidc.Provider)
			TenantAuth.verifier = make(map[string]*oidc.IDTokenVerifier)
		})
	}

}
