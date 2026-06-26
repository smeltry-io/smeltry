// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package tenant

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

var listKinds = map[schema.GroupVersionResource]string{
	nsGVR:    "NamespaceList",
	quotaGVR: "ResourceQuotaList",
}

func newFakeDyn(objs ...runtime.Object) *fake.FakeDynamicClient {
	scheme := runtime.NewScheme()
	return fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, objs...)
}

func makeNamespace(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata":   map[string]interface{}{"name": name},
	}}
}

func makeQuota(name, namespace, maxClusters, maxNodes string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ResourceQuota",
		"metadata":   map[string]interface{}{"name": name, "namespace": namespace},
		"spec": map[string]interface{}{
			"hard": map[string]interface{}{
				"count/clusterclaims.portal.smeltry.io": maxClusters,
				"count/serverclaims.portal.smeltry.io":  maxNodes,
			},
		},
	}}
}

func TestList_ReturnsTenantNamespaces(t *testing.T) {
	dyn := newFakeDyn(
		makeNamespace("tenant-acme"),
		makeNamespace("tenant-beta"),
		makeNamespace("kube-system"), // must be excluded
		makeQuota("default", "tenant-acme", "5", "20"),
	)

	tenants, err := List(context.Background(), dyn)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tenants) != 2 {
		t.Fatalf("expected 2 tenant namespaces, got %d", len(tenants))
	}
	names := map[string]bool{}
	for _, tn := range tenants {
		names[tn.Namespace] = true
	}
	if !names["tenant-acme"] {
		t.Error("expected tenant-acme in results")
	}
	if !names["tenant-beta"] {
		t.Error("expected tenant-beta in results")
	}
	if names["kube-system"] {
		t.Error("kube-system should be excluded")
	}
}

func TestList_QuotaExtracted(t *testing.T) {
	dyn := newFakeDyn(
		makeNamespace("tenant-acme"),
		makeQuota("default", "tenant-acme", "5", "20"),
	)

	tenants, err := List(context.Background(), dyn)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tenants) != 1 {
		t.Fatalf("expected 1 tenant, got %d", len(tenants))
	}
	if tenants[0].MaxClusters != "5" {
		t.Errorf("MaxClusters: got %q want %q", tenants[0].MaxClusters, "5")
	}
}

func TestList_NoTenants(t *testing.T) {
	dyn := newFakeDyn(makeNamespace("kube-system"))
	tenants, err := List(context.Background(), dyn)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tenants) != 0 {
		t.Errorf("expected 0 tenants, got %d", len(tenants))
	}
}
