// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package siteconfig provides read access to SiteConfig resources.
package siteconfig

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func strField(obj map[string]interface{}, keys ...string) string {
	cur := obj
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return ""
		}
		if i == len(keys)-1 {
			s, _ := v.(string)
			return s
		}
		cur, _ = v.(map[string]interface{})
		if cur == nil {
			return ""
		}
	}
	return ""
}

const systemNamespace = "portal-system"

var gvr = schema.GroupVersionResource{
	Group:    "portal.smeltry.io",
	Version:  "v1alpha1",
	Resource: "siteconfigs",
}

// SiteConfig is a minimal view of a SiteConfig resource.
type SiteConfig struct {
	Name string
}

// Client wraps a dynamic.Interface scoped to SiteConfig operations.
type Client struct {
	dyn dynamic.Interface
}

// NewClient returns a Client using the provided dynamic client.
func NewClient(dyn dynamic.Interface) *Client {
	return &Client{dyn: dyn}
}

// List returns all SiteConfigs from the portal-system namespace.
func (c *Client) List(ctx context.Context) ([]SiteConfig, error) {
	list, err := c.dyn.Resource(gvr).Namespace(systemNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing SiteConfigs: %w", err)
	}
	out := make([]SiteConfig, 0, len(list.Items))
	for _, item := range list.Items {
		out = append(out, SiteConfig{Name: strField(item.Object, "metadata", "name")})
	}
	return out, nil
}
