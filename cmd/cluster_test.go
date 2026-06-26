// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/smeltry-io/smeltry/internal/clusterclaim"
)

func TestPrintClusters_Table(t *testing.T) {
	items := []clusterclaim.ClusterClaim{
		{Name: "ml-train", Namespace: "tenant-acme", Phase: "Ready", Site: "paris-dc1", Class: "gpu-large", Count: 3, Age: "2h"},
		{Name: "build-01", Namespace: "tenant-acme", Phase: "Provisioning", Site: "paris-dc1", Class: "standard", Count: 1, Age: "5m"},
	}
	var buf bytes.Buffer
	cmd := newClusterListCmd()
	cmd.SetOut(&buf)

	if err := printClusters(cmd, items, "table"); err != nil {
		t.Fatalf("printClusters table: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"ml-train", "Ready", "gpu-large", "build-01", "Provisioning"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in table output, got:\n%s", want, out)
		}
	}
}

func TestPrintClusters_JSON(t *testing.T) {
	items := []clusterclaim.ClusterClaim{
		{Name: "ml-train", Phase: "Ready", Count: 3},
	}
	var buf bytes.Buffer
	cmd := newClusterListCmd()
	cmd.SetOut(&buf)

	if err := printClusters(cmd, items, "json"); err != nil {
		t.Fatalf("printClusters json: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"ml-train"`) {
		t.Errorf("expected JSON with ml-train, got: %s", out)
	}
}

func TestPrintCluster_Table(t *testing.T) {
	cc := clusterclaim.ClusterClaim{
		Name: "ml-train", Namespace: "tenant-acme", Phase: "Ready",
		Site: "paris-dc1", Class: "gpu-large", Count: 3, Age: "2h",
	}
	var buf bytes.Buffer
	cmd := newClusterGetCmd()
	cmd.SetOut(&buf)

	if err := printCluster(cmd, cc, "table"); err != nil {
		t.Fatalf("printCluster table: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"ml-train", "Ready", "paris-dc1", "gpu-large", "3"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
}

func TestRequireNamespace_Missing(t *testing.T) {
	old := global.Namespace
	global.Namespace = ""
	defer func() { global.Namespace = old }()

	cmd := newClusterListCmd()
	if err := requireNamespace(cmd, nil); err == nil {
		t.Error("expected error when namespace is missing, got nil")
	}
}

func TestRequireNamespace_Present(t *testing.T) {
	old := global.Namespace
	global.Namespace = "tenant-acme"
	defer func() { global.Namespace = old }()

	cmd := newClusterListCmd()
	if err := requireNamespace(cmd, nil); err != nil {
		t.Errorf("expected no error when namespace is set, got: %v", err)
	}
}

func TestWaitTimeout_Default(t *testing.T) {
	old := global.Timeout
	global.Timeout = ""
	defer func() { global.Timeout = old }()

	d, err := waitTimeout()
	if err != nil {
		t.Fatalf("waitTimeout default: %v", err)
	}
	if d != 10*time.Minute {
		t.Errorf("expected 10m default, got %v", d)
	}
}

func TestWaitTimeout_Custom(t *testing.T) {
	old := global.Timeout
	global.Timeout = "5m"
	defer func() { global.Timeout = old }()

	d, err := waitTimeout()
	if err != nil {
		t.Fatalf("waitTimeout custom: %v", err)
	}
	if d != 5*time.Minute {
		t.Errorf("expected 5m, got %v", d)
	}
}

func TestWaitTimeout_Invalid(t *testing.T) {
	old := global.Timeout
	global.Timeout = "not-a-duration"
	defer func() { global.Timeout = old }()

	_, err := waitTimeout()
	if err == nil {
		t.Error("expected error for invalid timeout, got nil")
	}
}

func TestClusterCreateFromFile_ValidManifest(t *testing.T) {
	// Write a minimal ClusterClaim manifest to a temp file.
	manifest := `
apiVersion: portal.smeltry.io/v1alpha1
kind: ClusterClaim
metadata:
  name: test-cluster
  namespace: tenant-acme
spec:
  site: paris-dc1
  machineClass: standard
  machineCount: 2
`
	f, err := os.CreateTemp(t.TempDir(), "cc-*.yaml")
	if err != nil {
		t.Fatalf("TempFile: %v", err)
	}
	if _, err := f.WriteString(manifest); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	f.Close()

	// clusterCreateFromFile calls k8sclient.New which needs a valid kubeconfig.
	// We only test the file parsing part by checking the error is not about YAML.
	// A real integration test would require a live cluster.
	err = clusterCreateFromFile(newClusterCreateCmd(), f.Name())
	// The error (if any) must be about the kube client, not about file parsing.
	if err != nil && strings.Contains(err.Error(), "parsing manifest") {
		t.Errorf("manifest parsing failed: %v", err)
	}
}

func TestClusterCreateFromFile_InvalidYAML(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "cc-*.yaml")
	if err != nil {
		t.Fatalf("TempFile: %v", err)
	}
	f.WriteString("{{invalid yaml{{")
	f.Close()

	err = clusterCreateFromFile(newClusterCreateCmd(), f.Name())
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestClusterCreateFromFile_MissingFile(t *testing.T) {
	err := clusterCreateFromFile(newClusterCreateCmd(), "/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestConfirmDelete_Confirmed(t *testing.T) {
	cmd := newClusterDeleteCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	r := strings.NewReader("ml-train\n")
	if !confirmResourceDelete(cmd.OutOrStdout(), r, "ClusterClaim", "ml-train", "tenant-acme") {
		t.Error("expected confirmDelete to return true when name matches")
	}
}

func TestConfirmDelete_Rejected(t *testing.T) {
	cmd := newClusterDeleteCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	r := strings.NewReader("wrong-name\n")
	if confirmResourceDelete(cmd.OutOrStdout(), r, "ClusterClaim", "ml-train", "tenant-acme") {
		t.Error("expected confirmDelete to return false when name does not match")
	}
}

func TestConfirmDelete_EmptyInput(t *testing.T) {
	cmd := newClusterDeleteCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	r := strings.NewReader("")
	if confirmResourceDelete(cmd.OutOrStdout(), r, "ClusterClaim", "ml-train", "tenant-acme") {
		t.Error("expected confirmDelete to return false on empty input (e.g. closed stdin)")
	}
}

func TestBoolStr(t *testing.T) {
	if boolStr(true) != "true" {
		t.Error("boolStr(true) should return \"true\"")
	}
	if boolStr(false) != "false" {
		t.Error("boolStr(false) should return \"false\"")
	}
}

func TestClusterCreateFromFile_NamespaceMismatch(t *testing.T) {
	manifest := `
apiVersion: portal.smeltry.io/v1alpha1
kind: ClusterClaim
metadata:
  name: test-cluster
  namespace: other-namespace
spec:
  site: paris-dc1
  machineClass: standard
  machineCount: 2
`
	f, err := os.CreateTemp(t.TempDir(), "cc-*.yaml")
	if err != nil {
		t.Fatalf("TempFile: %v", err)
	}
	f.WriteString(manifest)
	f.Close()

	old := global.Namespace
	global.Namespace = "tenant-acme"
	defer func() { global.Namespace = old }()

	cmd := newClusterCreateCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// We only care that a warning is printed, not about the kube error.
	_ = clusterCreateFromFile(cmd, f.Name())
	if !strings.Contains(buf.String(), "Warning") {
		t.Errorf("expected namespace mismatch warning, got: %q", buf.String())
	}
}
