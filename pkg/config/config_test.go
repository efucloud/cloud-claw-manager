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
	"bytes"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfig(t *testing.T) {
	config := &Config{

		LogConfig: &LogConfig{
			Level: "debug",
		},
		OidcConfig: OidcConfig{
			Issuer:       "https://eauth.efucloud.com",
			ClientId:     "umtsnwb4nvehzfmwaimi5knql",
			ClientSecret: "boowj4iujtvqinyw5grtmqzvpblcc6v7jhpxke5zhiktxn6zw3u",
		},
		OpenClawControl: OpenClawControlConfig{
			PreviewBaseDomain:    "openclawefucloud.com",
			AdminEmails:          []string{"admin@efucloud.com", "admin@efucloud.cn"},
			IngressEnabled:       true,
			IngressClassName:     "nginx",
			IngressPath:          "/",
			IngressPathType:      "Prefix",
			IngressTLSEnabled:    true,
			IngressTLSSecretName: "openclawefucloud-com-tls",
		},
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(config); err != nil {
		t.Fatalf("encode yaml: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("close encoder: %v", err)
	}
	if err := os.WriteFile("../../config/config.yaml", buf.Bytes(), os.ModePerm); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
}
