// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package clusterclaim provides read/write access to ClusterClaim resources
// via the Kubernetes dynamic client.
package clusterclaim

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var gvr = schema.GroupVersionResource{
	Group:    "portal.smeltry.io",
	Version:  "v1alpha1",
	Resource: "clusterclaims",
}

// ClusterClaim is a minimal representation of a ClusterClaim resource.
type ClusterClaim struct {
	Name      string
	Namespace string
	Phase     string
	Site      string
	Class     string
	Count     int64
	Age       string // human-readable, computed by caller
}

// Client wraps a dynamic.Interface scoped to ClusterClaim operations.
type Client struct {
	dyn dynamic.Interface
}

// NewClient returns a Client using the provided dynamic client.
func NewClient(dyn dynamic.Interface) *Client {
	return &Client{dyn: dyn}
}

// List returns all ClusterClaims in namespace.
func (c *Client) List(ctx context.Context, namespace string) ([]ClusterClaim, error) {
	list, err := c.dyn.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing ClusterClaims: %w", err)
	}
	out := make([]ClusterClaim, 0, len(list.Items))
	for _, item := range list.Items {
		out = append(out, fromUnstructured(item.Object))
	}
	return out, nil
}

// Get returns a single ClusterClaim by name.
func (c *Client) Get(ctx context.Context, namespace, name string) (*ClusterClaim, error) {
	obj, err := c.dyn.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting ClusterClaim %q: %w", name, err)
	}
	cc := fromUnstructured(obj.Object)
	return &cc, nil
}

// fromUnstructured converts raw unstructured data to a ClusterClaim.
func fromUnstructured(obj map[string]interface{}) ClusterClaim {
	cc := ClusterClaim{
		Name:      strField(obj, "metadata", "name"),
		Namespace: strField(obj, "metadata", "namespace"),
		Phase:     strField(obj, "status", "phase"),
		Site:      strField(obj, "spec", "site"),
		Class:     strField(obj, "spec", "machineClass"),
		Count:     intField(obj, "spec", "machineCount"),
	}
	return cc
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

func intField(obj map[string]interface{}, keys ...string) int64 {
	cur := obj
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return 0
		}
		if i == len(keys)-1 {
			switch n := v.(type) {
			case int64:
				return n
			case float64:
				return int64(n)
			}
			return 0
		}
		cur, _ = v.(map[string]interface{})
		if cur == nil {
			return 0
		}
	}
	return 0
}
