// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package helminstall

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildValues_SetOverridesFile(t *testing.T) {
	// Write a values file.
	dir := t.TempDir()
	vf := filepath.Join(dir, "vals.yaml")
	os.WriteFile(vf, []byte("replicaCount: 1\nimage:\n  tag: v0.1.0\n"), 0600)

	vals, err := buildValues([]string{vf}, []string{"replicaCount=3"})
	if err != nil {
		t.Fatalf("buildValues: %v", err)
	}
	if vals["replicaCount"] != int64(3) && vals["replicaCount"] != float64(3) {
		t.Errorf("expected replicaCount=3 (--set overrides file), got %v", vals["replicaCount"])
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

func TestChartRef_WithVersion(t *testing.T) {
	ref := chartRef("v1.2.3")
	if ref != DefaultChart {
		t.Errorf("chartRef should return DefaultChart regardless of version, got %q", ref)
	}
}

func TestChartRef_Latest(t *testing.T) {
	ref := chartRef("")
	if ref != DefaultChart {
		t.Errorf("chartRef(\"\") should return DefaultChart, got %q", ref)
	}
}
