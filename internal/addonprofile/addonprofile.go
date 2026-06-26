// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package addonprofile provides read access to AddonProfile resources.
package addonprofile

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const systemNamespace = "portal-system"

var gvr = schema.GroupVersionResource{
	Group:    "portal.smeltry.io",
	Version:  "v1alpha1",
	Resource: "addonprofiles",
}

// AddonProfile is a minimal view of an AddonProfile resource.
type AddonProfile struct {
	Name        string
	Description string
}

// Client wraps a dynamic.Interface scoped to AddonProfile operations.
type Client struct {
	dyn dynamic.Interface
}

// NewClient returns a Client using the provided dynamic client.
func NewClient(dyn dynamic.Interface) *Client {
	return &Client{dyn: dyn}
}

// List returns all AddonProfiles from the portal-system namespace.
func (c *Client) List(ctx context.Context) ([]AddonProfile, error) {
	list, err := c.dyn.Resource(gvr).Namespace(systemNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing AddonProfiles: %w", err)
	}
	out := make([]AddonProfile, 0, len(list.Items))
	for _, item := range list.Items {
		out = append(out, AddonProfile{
			Name:        strField(item.Object, "metadata", "name"),
			Description: strField(item.Object, "spec", "description"),
		})
	}
	return out, nil
}

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
