// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"bytes"
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
