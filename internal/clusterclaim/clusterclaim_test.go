// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package clusterclaim

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

// listKinds maps the ClusterClaim GVR to its list kind so the fake client can
// serve List requests correctly.
var listKinds = map[schema.GroupVersionResource]string{
	gvr: "ClusterClaimList",
}

func newFakeDyn(objs ...runtime.Object) *fake.FakeDynamicClient {
	scheme := runtime.NewScheme()
	return fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, objs...)
}

func makeCC(name, namespace, phase, site, class string, count int64) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "portal.smeltry.io/v1alpha1",
			"kind":       "ClusterClaim",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"site":         site,
				"machineClass": class,
				"machineCount": count,
			},
			"status": map[string]interface{}{
				"phase": phase,
			},
		},
	}
}

func TestList_ReturnsAllClaims(t *testing.T) {
	cc1 := makeCC("ml-train", "tenant-acme", "Ready", "paris-dc1", "gpu-large", 3)
	cc2 := makeCC("build-01", "tenant-acme", "Provisioning", "paris-dc1", "standard", 1)

	c := NewClient(newFakeDyn(cc1, cc2))
	list, err := c.List(context.Background(), "tenant-acme")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 ClusterClaims, got %d", len(list))
	}

	byName := map[string]ClusterClaim{}
	for _, cc := range list {
		byName[cc.Name] = cc
	}

	ml := byName["ml-train"]
	if ml.Phase != "Ready" {
		t.Errorf("ml-train phase: got %q want %q", ml.Phase, "Ready")
	}
	if ml.Count != 3 {
		t.Errorf("ml-train count: got %d want 3", ml.Count)
	}
	if ml.Class != "gpu-large" {
		t.Errorf("ml-train class: got %q want %q", ml.Class, "gpu-large")
	}
}

func TestList_EmptyNamespace(t *testing.T) {
	c := NewClient(newFakeDyn())
	list, err := c.List(context.Background(), "tenant-empty")
	if err != nil {
		t.Fatalf("List empty ns: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 results, got %d", len(list))
	}
}

func TestGet_ReturnsCorrectClaim(t *testing.T) {
	cc := makeCC("ml-train", "tenant-acme", "Ready", "paris-dc1", "gpu-large", 3)
	c := NewClient(newFakeDyn(cc))

	got, err := c.Get(context.Background(), "tenant-acme", "ml-train")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "ml-train" {
		t.Errorf("Name: got %q want %q", got.Name, "ml-train")
	}
	if got.Phase != "Ready" {
		t.Errorf("Phase: got %q want %q", got.Phase, "Ready")
	}
	if got.Site != "paris-dc1" {
		t.Errorf("Site: got %q want %q", got.Site, "paris-dc1")
	}
}

func TestGet_NotFound(t *testing.T) {
	c := NewClient(newFakeDyn())
	_, err := c.Get(context.Background(), "tenant-acme", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent ClusterClaim, got nil")
	}
}

func TestGet_WrongNamespace(t *testing.T) {
	cc := makeCC("ml-train", "tenant-acme", "Ready", "paris-dc1", "gpu-large", 3)
	c := NewClient(newFakeDyn(cc))

	_, err := c.Get(context.Background(), "tenant-other", "ml-train")
	if err == nil {
		t.Error("expected error when fetching from wrong namespace, got nil")
	}
}

func TestDelete_RemovesClaim(t *testing.T) {
	cc := makeCC("ml-train", "tenant-acme", "Ready", "paris-dc1", "gpu-large", 3)
	c := NewClient(newFakeDyn(cc))

	if err := c.Delete(context.Background(), "tenant-acme", "ml-train"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := c.Get(context.Background(), "tenant-acme", "ml-train")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestDelete_NotFound(t *testing.T) {
	c := NewClient(newFakeDyn())
	err := c.Delete(context.Background(), "tenant-acme", "nonexistent")
	if err == nil {
		t.Error("expected error deleting nonexistent ClusterClaim, got nil")
	}
}

func TestIsGone_TrueAfterDelete(t *testing.T) {
	cc := makeCC("ml-train", "tenant-acme", "Ready", "paris-dc1", "gpu-large", 3)
	c := NewClient(newFakeDyn(cc))

	if err := c.Delete(context.Background(), "tenant-acme", "ml-train"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	gone, err := c.IsGone(context.Background(), "tenant-acme", "ml-train")
	if err != nil {
		t.Fatalf("IsGone: %v", err)
	}
	if !gone {
		t.Error("expected IsGone=true after delete")
	}
}

func TestIsGone_FalseWhenPresent(t *testing.T) {
	cc := makeCC("ml-train", "tenant-acme", "Ready", "paris-dc1", "gpu-large", 3)
	c := NewClient(newFakeDyn(cc))

	gone, err := c.IsGone(context.Background(), "tenant-acme", "ml-train")
	if err != nil {
		t.Fatalf("IsGone: %v", err)
	}
	if gone {
		t.Error("expected IsGone=false for existing ClusterClaim")
	}
}

func TestHumanAge(t *testing.T) {
	tests := []struct {
		offset time.Duration
		want   string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m"},
		{3600 * time.Second, "1h"},
		{86400 * time.Second, "1d"},
	}
	for _, tt := range tests {
		got := humanAge(time.Now().Add(-tt.offset))
		if got != tt.want {
			t.Errorf("humanAge(-%v): got %q want %q", tt.offset, got, tt.want)
		}
	}
}

func TestCreate_ReturnsNewClaim(t *testing.T) {
	c := NewClient(newFakeDyn())
	obj := map[string]interface{}{
		"apiVersion": "portal.smeltry.io/v1alpha1",
		"kind":       "ClusterClaim",
		"metadata":   map[string]interface{}{"name": "new-cluster", "namespace": "tenant-acme"},
		"spec":       map[string]interface{}{"site": "paris-dc1", "machineClass": "standard", "machineCount": int64(2)},
	}
	got, err := c.Create(context.Background(), "tenant-acme", obj)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.Name != "new-cluster" {
		t.Errorf("Name: got %q want %q", got.Name, "new-cluster")
	}
}
