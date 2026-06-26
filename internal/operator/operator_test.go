// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package operator

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

var listKinds = map[schema.GroupVersionResource]string{
	deployGVR: "DeploymentList",
	ccGVR:     "ClusterClaimList",
	scGVR:     "ServerClaimList",
}

func newFakeDyn(objs ...runtime.Object) *fake.FakeDynamicClient {
	scheme := runtime.NewScheme()
	return fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, objs...)
}

func makeDeployment(name, namespace, version string, replicas, ready int64) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
			"labels": map[string]interface{}{
				"app.kubernetes.io/version": version,
			},
		},
		"spec": map[string]interface{}{
			"replicas": replicas,
		},
		"status": map[string]interface{}{
			"replicas":      replicas,
			"readyReplicas": ready,
		},
	}}
}

func makeCC(name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "portal.smeltry.io/v1alpha1",
		"kind":       "ClusterClaim",
		"metadata":   map[string]interface{}{"name": name, "namespace": namespace},
	}}
}

func makeSC(name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "portal.smeltry.io/v1alpha1",
		"kind":       "ServerClaim",
		"metadata":   map[string]interface{}{"name": name, "namespace": namespace},
	}}
}

func TestFetch_OperatorReady(t *testing.T) {
	dyn := newFakeDyn(
		makeDeployment("smeltry-operator", "smeltry-system", "v0.1.0", 1, 1),
		makeCC("cc1", "tenant-a"),
		makeCC("cc2", "tenant-b"),
		makeSC("sc1", "tenant-a"),
	)
	st, err := Fetch(context.Background(), dyn)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !st.Ready {
		t.Error("expected Ready=true")
	}
	if st.Version != "v0.1.0" {
		t.Errorf("Version: got %q want %q", st.Version, "v0.1.0")
	}
	if st.ClusterClaims != 2 {
		t.Errorf("ClusterClaims: got %d want 2", st.ClusterClaims)
	}
	if st.ServerClaims != 1 {
		t.Errorf("ServerClaims: got %d want 1", st.ServerClaims)
	}
}

func TestFetch_OperatorNotReady(t *testing.T) {
	dyn := newFakeDyn(
		makeDeployment("smeltry-operator", "smeltry-system", "v0.1.0", 1, 0),
	)
	st, err := Fetch(context.Background(), dyn)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if st.Ready {
		t.Error("expected Ready=false when readyReplicas=0")
	}
	if st.Absent {
		t.Error("expected Absent=false when deployment exists")
	}
}

func TestFetch_OperatorAbsent(t *testing.T) {
	dyn := newFakeDyn()
	st, err := Fetch(context.Background(), dyn)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !st.Absent {
		t.Error("expected Absent=true when deployment not found")
	}
	if st.Ready {
		t.Error("expected Ready=false when absent")
	}
}

func TestFetch_CountsAcrossNamespaces(t *testing.T) {
	dyn := newFakeDyn(
		makeDeployment("smeltry-operator", "smeltry-system", "v0.1.0", 1, 1),
		makeCC("cc1", "tenant-a"),
		makeCC("cc2", "tenant-b"),
		makeCC("cc3", "tenant-c"),
		makeSC("sc1", "tenant-a"),
		makeSC("sc2", "tenant-b"),
	)
	st, err := Fetch(context.Background(), dyn)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if st.ClusterClaims != 3 {
		t.Errorf("ClusterClaims: got %d want 3", st.ClusterClaims)
	}
	if st.ServerClaims != 2 {
		t.Errorf("ServerClaims: got %d want 2", st.ServerClaims)
	}
}
