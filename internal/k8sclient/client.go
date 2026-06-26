// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package k8sclient builds a REST client authenticated with the stored OIDC token.
package k8sclient

import (
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/smeltry-io/smeltry/internal/auth"
)

// New returns a dynamic client using the stored OIDC token as Bearer.
// It resolves the server URL from KUBECONFIG / in-cluster config, then
// overrides the auth with the smeltry token.
func New(serverOverride string) (dynamic.Interface, error) {
	td, err := auth.Load()
	if err != nil {
		return nil, err
	}

	cfg, err := baseConfig(serverOverride)
	if err != nil {
		return nil, fmt.Errorf("resolving kubeconfig: %w", err)
	}

	cfg.BearerToken = td.AccessToken
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
