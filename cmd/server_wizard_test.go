// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/smeltry-io/smeltry/internal/siteconfig"
)

func TestServerWizard_HappyPath(t *testing.T) {
	sites := []siteconfig.SiteConfig{{Name: "paris-dc1"}, {Name: "london-dc1"}}
	// Input: pick site 1, machineClass "gpu-large", OS "flatcar", name "build-01"
	input := "1\ngpu-large\n1\nbuild-01\n"
	var out bytes.Buffer
	spec, err := runServerWizard(&out, strings.NewReader(input), sites)
	if err != nil {
		t.Fatalf("runServerWizard: %v", err)
	}
	if spec.site != "paris-dc1" {
		t.Errorf("site: got %q want %q", spec.site, "paris-dc1")
	}
	if spec.machineClass != "gpu-large" {
		t.Errorf("machineClass: got %q want %q", spec.machineClass, "gpu-large")
	}
	if spec.os != "flatcar" {
		t.Errorf("os: got %q want %q", spec.os, "flatcar")
	}
	if spec.name != "build-01" {
		t.Errorf("name: got %q want %q", spec.name, "build-01")
	}
}

func TestServerWizard_Abort(t *testing.T) {
	sites := []siteconfig.SiteConfig{{Name: "paris-dc1"}}
	input := "q\n"
	var out bytes.Buffer
	_, err := runServerWizard(&out, strings.NewReader(input), sites)
	if err == nil {
		t.Fatal("expected error on abort")
	}
}

func TestServerWizard_NoSites(t *testing.T) {
	var out bytes.Buffer
	_, err := runServerWizard(&out, strings.NewReader(""), nil)
	if err == nil {
		t.Fatal("expected error when no sites are available")
	}
}

func TestServerWizard_SecondSite(t *testing.T) {
	sites := []siteconfig.SiteConfig{{Name: "paris-dc1"}, {Name: "london-dc1"}}
	input := "2\nstandard\n2\nmy-server\n"
	var out bytes.Buffer
	spec, err := runServerWizard(&out, strings.NewReader(input), sites)
	if err != nil {
		t.Fatalf("runServerWizard: %v", err)
	}
	if spec.site != "london-dc1" {
		t.Errorf("site: got %q want %q", spec.site, "london-dc1")
	}
	if spec.os != "ubuntu" {
		t.Errorf("os: got %q want %q", spec.os, "ubuntu")
	}
}

func TestServerSpecToObject_Structure(t *testing.T) {
	spec := serverSpec{
		name:         "build-01",
		site:         "paris-dc1",
		machineClass: "standard",
		os:           "flatcar",
	}
	obj := serverSpecToObject("tenant-acme", spec)

	meta, _ := obj["metadata"].(map[string]interface{})
	if meta["name"] != "build-01" {
		t.Errorf("metadata.name: got %v", meta["name"])
	}
	if meta["namespace"] != "tenant-acme" {
		t.Errorf("metadata.namespace: got %v", meta["namespace"])
	}

	sp, _ := obj["spec"].(map[string]interface{})
	if sp["site"] != "paris-dc1" {
		t.Errorf("spec.site: got %v", sp["site"])
	}
	if sp["os"] != "flatcar" {
		t.Errorf("spec.os: got %v", sp["os"])
	}
}
