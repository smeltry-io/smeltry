// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package serverclaim

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

var listKinds = map[schema.GroupVersionResource]string{
	gvr: "ServerClaimList",
}

func newFakeDyn(objs ...runtime.Object) *fake.FakeDynamicClient {
	scheme := runtime.NewScheme()
	return fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, objs...)
}

func makeServerClaim(name, namespace, phase, site, class, os string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "portal.smeltry.io/v1alpha1",
		"kind":       "ServerClaim",
		"metadata": map[string]interface{}{
			"name":              name,
			"namespace":         namespace,
			"creationTimestamp": time.Now().UTC().Format(time.RFC3339),
		},
		"spec": map[string]interface{}{
			"site":         site,
			"machineClass": class,
			"os":           os,
		},
		"status": map[string]interface{}{
			"phase": phase,
		},
	}}
}

func TestList_ReturnsClaims(t *testing.T) {
	obj := makeServerClaim("build-01", "tenant-acme", "Ready", "paris-dc1", "standard", "flatcar")
	dyn := newFakeDyn(obj)

	items, err := NewClient(dyn).List(context.Background(), "tenant-acme")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	sc := items[0]
	if sc.Name != "build-01" {
		t.Errorf("Name: got %q want %q", sc.Name, "build-01")
	}
	if sc.Phase != "Ready" {
		t.Errorf("Phase: got %q want %q", sc.Phase, "Ready")
	}
	if sc.Site != "paris-dc1" {
		t.Errorf("Site: got %q want %q", sc.Site, "paris-dc1")
	}
	if sc.Class != "standard" {
		t.Errorf("Class: got %q want %q", sc.Class, "standard")
	}
	if sc.OS != "flatcar" {
		t.Errorf("OS: got %q want %q", sc.OS, "flatcar")
	}
	if sc.Age == "" {
		t.Error("Age should be set")
	}
}

func TestList_EmptyNamespace(t *testing.T) {
	obj := makeServerClaim("build-01", "tenant-acme", "Ready", "paris-dc1", "standard", "flatcar")
	dyn := newFakeDyn(obj)

	items, err := NewClient(dyn).List(context.Background(), "tenant-other")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items in different namespace, got %d", len(items))
	}
}

func TestGet_ReturnsServerClaim(t *testing.T) {
	obj := makeServerClaim("build-01", "tenant-acme", "Provisioning", "paris-dc1", "gpu-large", "ubuntu")
	dyn := newFakeDyn(obj)

	sc, err := NewClient(dyn).Get(context.Background(), "tenant-acme", "build-01")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if sc.Name != "build-01" {
		t.Errorf("Name: got %q want %q", sc.Name, "build-01")
	}
	if sc.OS != "ubuntu" {
		t.Errorf("OS: got %q want %q", sc.OS, "ubuntu")
	}
	if sc.Class != "gpu-large" {
		t.Errorf("Class: got %q want %q", sc.Class, "gpu-large")
	}
}

func TestGet_NotFound(t *testing.T) {
	dyn := newFakeDyn()

	_, err := NewClient(dyn).Get(context.Background(), "tenant-acme", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent ServerClaim")
	}
}

func TestDelete_RemovesClaim(t *testing.T) {
	obj := makeServerClaim("build-01", "tenant-acme", "Ready", "paris-dc1", "standard", "flatcar")
	dyn := newFakeDyn(obj)

	if err := NewClient(dyn).Delete(context.Background(), "tenant-acme", "build-01"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := dyn.Resource(gvr).Namespace("tenant-acme").Get(context.Background(), "build-01", metav1.GetOptions{})
	if err == nil {
		t.Error("expected Not Found after Delete")
	}
}

func TestIsGone_TrueWhenDeleted(t *testing.T) {
	dyn := newFakeDyn()

	gone, err := NewClient(dyn).IsGone(context.Background(), "tenant-acme", "build-01")
	if err != nil {
		t.Fatalf("IsGone: %v", err)
	}
	if !gone {
		t.Error("expected gone=true for missing resource")
	}
}

func TestIsGone_FalseWhenPresent(t *testing.T) {
	obj := makeServerClaim("build-01", "tenant-acme", "Ready", "paris-dc1", "standard", "flatcar")
	dyn := newFakeDyn(obj)

	gone, err := NewClient(dyn).IsGone(context.Background(), "tenant-acme", "build-01")
	if err != nil {
		t.Fatalf("IsGone: %v", err)
	}
	if gone {
		t.Error("expected gone=false for existing resource")
	}
}

func TestCreate_ReturnsServerClaim(t *testing.T) {
	dyn := newFakeDyn()

	obj := map[string]interface{}{
		"apiVersion": "portal.smeltry.io/v1alpha1",
		"kind":       "ServerClaim",
		"metadata": map[string]interface{}{
			"name":      "build-02",
			"namespace": "tenant-acme",
		},
		"spec": map[string]interface{}{
			"site":         "paris-dc1",
			"machineClass": "standard",
			"os":           "flatcar",
		},
	}

	sc, err := NewClient(dyn).Create(context.Background(), "tenant-acme", obj)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sc.Name != "build-02" {
		t.Errorf("Name: got %q want %q", sc.Name, "build-02")
	}
}

func TestFromUnstructured_AllFields(t *testing.T) {
	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":              "srv-01",
			"namespace":         "tenant-acme",
			"creationTimestamp": time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339),
		},
		"spec": map[string]interface{}{
			"site":         "paris-dc1",
			"machineClass": "standard",
			"os":           "flatcar",
		},
		"status": map[string]interface{}{
			"phase":    "Ready",
			"serverIP": "10.0.1.42",
		},
	}
	sc := fromUnstructured(obj)
	if sc.Phase != "Ready" {
		t.Errorf("Phase: %q", sc.Phase)
	}
	if sc.ServerIP != "10.0.1.42" {
		t.Errorf("ServerIP: %q", sc.ServerIP)
	}
	if sc.Age == "" {
		t.Error("Age should not be empty")
	}
}

func TestFromUnstructured_MissingStatus(t *testing.T) {
	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "srv-01",
			"namespace": "tenant-acme",
		},
		"spec": map[string]interface{}{
			"site": "paris-dc1",
			"os":   "flatcar",
		},
	}
	sc := fromUnstructured(obj)
	if sc.Phase != "" {
		t.Errorf("Phase should be empty, got %q", sc.Phase)
	}
	if sc.ServerIP != "" {
		t.Errorf("ServerIP should be empty, got %q", sc.ServerIP)
	}
}
