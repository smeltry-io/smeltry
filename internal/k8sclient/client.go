// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package k8sclient builds a REST client authenticated with the stored OIDC token.
package k8sclient

import (
	"fmt"
	"os"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/smeltry-io/smeltry/internal/auth"
)

// EnvToken mirrors cmd.EnvToken to avoid an import cycle.
// Both constants must remain identical.
const envToken = "SMELTRY_TOKEN"

// New returns a dynamic client using the OIDC Bearer token.
// Token resolution order:
//  1. SMELTRY_TOKEN environment variable (CI — nothing written to disk)
//  2. Stored token from ~/.config/smeltry/token.json
//
// The server URL is resolved from KUBECONFIG / in-cluster config, then
// optionally overridden by serverOverride.
func New(serverOverride string) (dynamic.Interface, error) {
	bearer := os.Getenv(envToken)
	if bearer == "" {
		td, err := auth.Load()
		if err != nil {
			return nil, err
		}
		bearer = td.AccessToken
	}

	cfg, err := baseConfig(serverOverride)
	if err != nil {
		return nil, fmt.Errorf("resolving kubeconfig: %w", err)
	}

	cfg.BearerToken = bearer
	cfg.BearerTokenFile = ""
	// Clear any credential provider set by kubeconfig; the Bearer token is enough.
	cfg.Username = ""
	cfg.Password = ""
	cfg.TLSClientConfig.CertFile = ""
	cfg.TLSClientConfig.KeyFile = ""

	return dynamic.NewForConfig(cfg)
}

// baseConfig loads the server address from KUBECONFIG or in-cluster config.
// If serverOverride is non-empty it takes precedence.
func baseConfig(serverOverride string) (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	if serverOverride != "" {
		overrides.ClusterInfo.Server = serverOverride
	}
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
	return cc.ClientConfig()
}
