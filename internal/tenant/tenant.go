// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package tenant lists Kubernetes namespaces that represent Smeltry tenants
// (those with the "tenant-" prefix) and their associated ResourceQuota.
package tenant

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	nsGVR    = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	quotaGVR = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "resourcequotas"}
)

const tenantPrefix = "tenant-"

// Tenant holds the aggregated view of a tenant namespace.
type Tenant struct {
	Namespace   string
	MaxClusters string
	MaxNodes    string
}

// List returns all tenant namespaces and their quota values.
func List(ctx context.Context, dyn dynamic.Interface) ([]Tenant, error) {
	nsList, err := dyn.Resource(nsGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing namespaces: %w", err)
	}

	var tenants []Tenant
	for _, ns := range nsList.Items {
		name, _, _ := unstructured.NestedString(ns.Object, "metadata", "name")
		if !strings.HasPrefix(name, tenantPrefix) {
			continue
		}

		tn := Tenant{Namespace: name}

		quotas, err := dyn.Resource(quotaGVR).Namespace(name).List(ctx, metav1.ListOptions{})
		if err == nil && len(quotas.Items) > 0 {
			hard, _, _ := unstructured.NestedMap(quotas.Items[0].Object, "spec", "hard")
			tn.MaxClusters, _ = hard["count/clusterclaims.portal.smeltry.io"].(string)
			tn.MaxNodes, _ = hard["count/serverclaims.portal.smeltry.io"].(string)
		}

		tenants = append(tenants, tn)
	}
	return tenants, nil
}
