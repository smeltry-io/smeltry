// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package addon provides read access to capi-addon-provider HelmRelease objects.
package addon

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// HelmRelease GVR for capi-addon-provider (addons.stackhpc.com/v1alpha1).
var gvr = schema.GroupVersionResource{
	Group:    "addons.stackhpc.com",
	Version:  "v1alpha1",
	Resource: "helmreleases",
}

// HelmRelease is a minimal view of a capi-addon-provider HelmRelease.
type HelmRelease struct {
	Name        string
	Namespace   string
	ClusterName string
	Ready       bool
	Bootstrap   bool
}

// Client wraps a dynamic.Interface scoped to HelmRelease operations.
type Client struct {
	dyn dynamic.Interface
}

// NewClient returns a Client using the provided dynamic client.
func NewClient(dyn dynamic.Interface) *Client {
	return &Client{dyn: dyn}
}

// ListForCluster returns all HelmReleases in namespace whose spec.clusterName
// matches clusterName.
func (c *Client) ListForCluster(ctx context.Context, namespace, clusterName string) ([]HelmRelease, error) {
	list, err := c.dyn.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing HelmReleases: %w", err)
	}
	var out []HelmRelease
	for _, item := range list.Items {
		hr := fromUnstructured(item.Object)
		if hr.ClusterName == clusterName {
			out = append(out, hr)
		}
	}
	return out, nil
}

func fromUnstructured(obj map[string]interface{}) HelmRelease {
	return HelmRelease{
		Name:        strField(obj, "metadata", "name"),
		Namespace:   strField(obj, "metadata", "namespace"),
		ClusterName: strField(obj, "spec", "clusterName"),
		Ready:       boolField(obj, "status", "ready"),
		Bootstrap:   boolField(obj, "spec", "bootstrap"),
	}
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

func boolField(obj map[string]interface{}, keys ...string) bool {
	cur := obj
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return false
		}
		if i == len(keys)-1 {
			b, _ := v.(bool)
			return b
		}
		cur, _ = v.(map[string]interface{})
		if cur == nil {
			return false
		}
	}
	return false
}
