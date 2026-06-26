// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package addonprofile

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

var listKinds = map[schema.GroupVersionResource]string{
	gvr: "AddonProfileList",
}

func newFakeDyn(objs ...runtime.Object) *fake.FakeDynamicClient {
	return fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, objs...)
}

func makeAddonProfile(name, description string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "portal.smeltry.io/v1alpha1",
			"kind":       "AddonProfile",
			"metadata":   map[string]interface{}{"name": name, "namespace": "portal-system"},
			"spec":       map[string]interface{}{"description": description},
		},
	}
}

func TestList_ReturnsProfiles(t *testing.T) {
	ap1 := makeAddonProfile("gpu-compute", "Cilium + GPU operator")
	ap2 := makeAddonProfile("standard", "Cilium + Ingress")

	c := NewClient(newFakeDyn(ap1, ap2))
	profiles, err := c.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 AddonProfiles, got %d", len(profiles))
	}
}

func TestList_Empty(t *testing.T) {
	c := NewClient(newFakeDyn())
	profiles, err := c.List(context.Background())
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0 results, got %d", len(profiles))
	}
}

func TestList_FieldsPreserved(t *testing.T) {
	ap := makeAddonProfile("gpu-compute", "Cilium + GPU operator")
	c := NewClient(newFakeDyn(ap))

	profiles, err := c.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if profiles[0].Name != "gpu-compute" {
		t.Errorf("Name: got %q want %q", profiles[0].Name, "gpu-compute")
	}
	if profiles[0].Description != "Cilium + GPU operator" {
		t.Errorf("Description: got %q want %q", profiles[0].Description, "Cilium + GPU operator")
	}
}
