// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package siteconfig

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

var listKinds = map[schema.GroupVersionResource]string{
	gvr: "SiteConfigList",
}

func newFakeDyn(objs ...runtime.Object) *fake.FakeDynamicClient {
	return fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, objs...)
}

func makeSiteConfig(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "portal.smeltry.io/v1alpha1",
			"kind":       "SiteConfig",
			"metadata":   map[string]interface{}{"name": name, "namespace": "portal-system"},
		},
	}
}

func TestList_ReturnsSites(t *testing.T) {
	sc1 := makeSiteConfig("paris-dc1")
	sc2 := makeSiteConfig("london-dc1")

	c := NewClient(newFakeDyn(sc1, sc2))
	sites, err := c.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sites) != 2 {
		t.Fatalf("expected 2 SiteConfigs, got %d", len(sites))
	}
}

func TestList_Empty(t *testing.T) {
	c := NewClient(newFakeDyn())
	sites, err := c.List(context.Background())
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(sites) != 0 {
		t.Errorf("expected 0 results, got %d", len(sites))
	}
}

func TestList_NamesPreserved(t *testing.T) {
	sc := makeSiteConfig("paris-dc1")
	c := NewClient(newFakeDyn(sc))

	sites, err := c.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if sites[0].Name != "paris-dc1" {
		t.Errorf("expected name %q, got %q", "paris-dc1", sites[0].Name)
	}
}
