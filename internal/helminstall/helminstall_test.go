// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package helminstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildValues_SetOverridesFile(t *testing.T) {
	dir := t.TempDir()
	vf := filepath.Join(dir, "vals.yaml")
	// File sets replicaCount=1; --set must win with replicaCount=3.
	os.WriteFile(vf, []byte("replicaCount: 1\nimage:\n  tag: v0.1.0\n"), 0600)

	vals, err := buildValues([]string{vf}, []string{"replicaCount=3"})
	if err != nil {
		t.Fatalf("buildValues: %v", err)
	}
	got := vals["replicaCount"]
	if got != int64(3) && got != float64(3) {
		t.Errorf("--set should override file: want replicaCount=3, got %v", got)
	}
	// Sanity-check: image.tag from the file is still present.
	img, _ := vals["image"].(map[string]interface{})
	if img["tag"] != "v0.1.0" {
		t.Errorf("file value image.tag should be preserved, got %v", img["tag"])
	}
}

func TestBuildValues_FileOnly(t *testing.T) {
	dir := t.TempDir()
	vf := filepath.Join(dir, "vals.yaml")
	os.WriteFile(vf, []byte("operator:\n  logLevel: debug\n"), 0600)

	vals, err := buildValues([]string{vf}, nil)
	if err != nil {
		t.Fatalf("buildValues: %v", err)
	}
	op, _ := vals["operator"].(map[string]interface{})
	if op["logLevel"] != "debug" {
		t.Errorf("expected logLevel=debug, got %v", op["logLevel"])
	}
}

func TestBuildValues_SetOnly(t *testing.T) {
	vals, err := buildValues(nil, []string{"foo=bar", "count=5"})
	if err != nil {
		t.Fatalf("buildValues: %v", err)
	}
	if vals["foo"] != "bar" {
		t.Errorf("foo: got %v want bar", vals["foo"])
	}
}

func TestBuildValues_InvalidFile(t *testing.T) {
	_, err := buildValues([]string{"/nonexistent/file.yaml"}, nil)
	if err == nil {
		t.Error("expected error for missing values file")
	}
}

func TestBuildValues_Empty(t *testing.T) {
	vals, err := buildValues(nil, nil)
	if err != nil {
		t.Fatalf("buildValues empty: %v", err)
	}
	if len(vals) != 0 {
		t.Errorf("expected empty map, got %v", vals)
	}
}

func TestDefaultChart_IsOCIRef(t *testing.T) {
	if !strings.HasPrefix(DefaultChart, "oci://") {
		t.Errorf("DefaultChart should be an OCI reference, got %q", DefaultChart)
	}
}
