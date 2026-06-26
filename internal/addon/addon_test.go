// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package addon

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

var listKinds = map[schema.GroupVersionResource]string{
	gvr: "HelmReleaseList",
}

func newFakeDyn(objs ...runtime.Object) *fake.FakeDynamicClient {
	return fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, objs...)
}

func makeHR(name, namespace, clusterName string, ready, bootstrap bool) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "addons.stackhpc.com/v1alpha1",
			"kind":       "HelmRelease",
			"metadata":   map[string]interface{}{"name": name, "namespace": namespace},
			"spec": map[string]interface{}{
				"clusterName": clusterName,
				"bootstrap":   bootstrap,
			},
			"status": map[string]interface{}{
				"ready": ready,
			},
		},
	}
}

func TestListForCluster_FiltersCorrectly(t *testing.T) {
	hr1 := makeHR("cilium", "tenant-acme", "ml-train", true, true)
	hr2 := makeHR("ingress", "tenant-acme", "ml-train", false, false)
	hr3 := makeHR("rook-ceph", "tenant-acme", "other-cluster", true, false) // different name + different cluster

	c := NewClient(newFakeDyn(hr1, hr2, hr3))
	list, err := c.ListForCluster(context.Background(), "tenant-acme", "ml-train")
	if err != nil {
		t.Fatalf("ListForCluster: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 HelmReleases for ml-train, got %d", len(list))
	}
	for _, hr := range list {
		if hr.ClusterName != "ml-train" {
			t.Errorf("unexpected clusterName %q", hr.ClusterName)
		}
	}
}

func TestListForCluster_ReadyAndBootstrapFields(t *testing.T) {
	hr := makeHR("cilium", "tenant-acme", "ml-train", true, true)
	c := NewClient(newFakeDyn(hr))

	list, err := c.ListForCluster(context.Background(), "tenant-acme", "ml-train")
	if err != nil {
		t.Fatalf("ListForCluster: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 result, got %d", len(list))
	}
	if !list[0].Ready {
		t.Error("expected Ready=true")
	}
	if !list[0].Bootstrap {
		t.Error("expected Bootstrap=true")
	}
}

func TestListForCluster_Empty(t *testing.T) {
	c := NewClient(newFakeDyn())
	list, err := c.ListForCluster(context.Background(), "tenant-acme", "ml-train")
	if err != nil {
		t.Fatalf("ListForCluster empty: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 results, got %d", len(list))
	}
}
