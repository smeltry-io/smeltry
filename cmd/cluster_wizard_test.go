// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/smeltry-io/smeltry/internal/addonprofile"
	"github.com/smeltry-io/smeltry/internal/siteconfig"
)

func TestWizard_CollectsAllFields(t *testing.T) {
	sites := []siteconfig.SiteConfig{{Name: "paris-dc1"}, {Name: "london-dc1"}}
	profiles := []addonprofile.AddonProfile{
		{Name: "gpu-compute", Description: "Cilium + GPU operator"},
		{Name: "standard", Description: "Cilium + Ingress"},
	}

	// Input: name, site choice (1), machineClass, machineCount, addonProfile choice (2), confirm
	input := "ml-train\n1\ngpu-large\n3\n2\ny\n"
	var out bytes.Buffer

	spec, err := runWizard(&out, strings.NewReader(input), sites, profiles)
	if err != nil {
		t.Fatalf("runWizard: %v", err)
	}
	if spec.Name != "ml-train" {
		t.Errorf("Name: got %q want %q", spec.Name, "ml-train")
	}
	if spec.Site != "paris-dc1" {
		t.Errorf("Site: got %q want %q", spec.Site, "paris-dc1")
	}
	if spec.MachineClass != "gpu-large" {
		t.Errorf("MachineClass: got %q want %q", spec.MachineClass, "gpu-large")
	}
	if spec.MachineCount != 3 {
		t.Errorf("MachineCount: got %d want 3", spec.MachineCount)
	}
	if spec.AddonProfile != "standard" {
		t.Errorf("AddonProfile: got %q want %q", spec.AddonProfile, "standard")
	}
}

func TestWizard_AbortOnNoConfirm(t *testing.T) {
	sites := []siteconfig.SiteConfig{{Name: "paris-dc1"}}
	profiles := []addonprofile.AddonProfile{{Name: "standard", Description: "Cilium + Ingress"}}

	input := "ml-train\n1\ngpu-large\n1\n1\nN\n"
	var out bytes.Buffer

	_, err := runWizard(&out, strings.NewReader(input), sites, profiles)
	if err != errWizardAborted {
		t.Errorf("expected errWizardAborted, got %v", err)
	}
}

func TestWizard_DefaultMachineCount(t *testing.T) {
	sites := []siteconfig.SiteConfig{{Name: "paris-dc1"}}
	profiles := []addonprofile.AddonProfile{{Name: "standard", Description: "Cilium + Ingress"}}

	// Empty machineCount input → default 1
	input := "ml-train\n1\ngpu-large\n\n1\ny\n"
	var out bytes.Buffer

	spec, err := runWizard(&out, strings.NewReader(input), sites, profiles)
	if err != nil {
		t.Fatalf("runWizard: %v", err)
	}
	if spec.MachineCount != 1 {
		t.Errorf("MachineCount default: got %d want 1", spec.MachineCount)
	}
}

func TestWizard_InvalidChoiceFallsBackToFirst(t *testing.T) {
	sites := []siteconfig.SiteConfig{{Name: "paris-dc1"}, {Name: "london-dc1"}}
	profiles := []addonprofile.AddonProfile{{Name: "standard", Description: "Cilium + Ingress"}}

	// Site choice "99" is out of range → first item used
	input := "ml-train\n99\ngpu-large\n1\n1\ny\n"
	var out bytes.Buffer

	spec, err := runWizard(&out, strings.NewReader(input), sites, profiles)
	if err != nil {
		t.Fatalf("runWizard: %v", err)
	}
	if spec.Site != "paris-dc1" {
		t.Errorf("Site fallback: got %q want %q", spec.Site, "paris-dc1")
	}
}

func TestWizard_NoSitesAvailable(t *testing.T) {
	var out bytes.Buffer
	_, err := runWizard(&out, strings.NewReader(""), nil, []addonprofile.AddonProfile{{Name: "standard"}})
	if err == nil {
		t.Error("expected error when no sites available")
	}
}

func TestWizard_NoProfilesAvailable(t *testing.T) {
	var out bytes.Buffer
	_, err := runWizard(&out, strings.NewReader(""), []siteconfig.SiteConfig{{Name: "paris-dc1"}}, nil)
	if err == nil {
		t.Error("expected error when no addon profiles available")
	}
}

func TestWizard_EmptyNameRejected(t *testing.T) {
	sites := []siteconfig.SiteConfig{{Name: "paris-dc1"}}
	profiles := []addonprofile.AddonProfile{{Name: "standard", Description: "Cilium + Ingress"}}

	// Empty name → should error immediately
	input := "\n"
	var out bytes.Buffer

	_, err := runWizard(&out, strings.NewReader(input), sites, profiles)
	if err == nil {
		t.Error("expected error for empty cluster name")
	}
}

func TestWizard_EmptyMachineClassRejected(t *testing.T) {
	sites := []siteconfig.SiteConfig{{Name: "paris-dc1"}}
	profiles := []addonprofile.AddonProfile{{Name: "standard", Description: "Cilium + Ingress"}}

	// Valid name and site, empty machine class → should error
	input := "ml-train\n1\n\n"
	var out bytes.Buffer

	_, err := runWizard(&out, strings.NewReader(input), sites, profiles)
	if err == nil {
		t.Error("expected error for empty machine class")
	}
}

func TestWizard_InvalidProfileChoiceFallsBackToFirst(t *testing.T) {
	sites := []siteconfig.SiteConfig{{Name: "paris-dc1"}}
	profiles := []addonprofile.AddonProfile{
		{Name: "standard", Description: "Cilium + Ingress"},
		{Name: "gpu-compute", Description: "Cilium + GPU operator"},
	}

	// Profile choice "99" is out of range → first item used
	input := "ml-train\n1\ngpu-large\n1\n99\ny\n"
	var out bytes.Buffer

	spec, err := runWizard(&out, strings.NewReader(input), sites, profiles)
	if err != nil {
		t.Fatalf("runWizard: %v", err)
	}
	if spec.AddonProfile != "standard" {
		t.Errorf("AddonProfile fallback: got %q want %q", spec.AddonProfile, "standard")
	}
}

func TestClusterSpecToObject(t *testing.T) {
	spec := clusterSpec{
		Name:         "ml-train",
		Site:         "paris-dc1",
		MachineClass: "gpu-large",
		MachineCount: 3,
		AddonProfile: "gpu-compute",
	}
	obj := clusterSpecToObject("tenant-acme", spec)

	if got, _ := obj["apiVersion"].(string); got != "portal.smeltry.io/v1alpha1" {
		t.Errorf("apiVersion: got %q want %q", got, "portal.smeltry.io/v1alpha1")
	}
	if got, _ := obj["kind"].(string); got != "ClusterClaim" {
		t.Errorf("kind: got %q want %q", got, "ClusterClaim")
	}

	meta, _ := obj["metadata"].(map[string]interface{})
	if meta["name"] != "ml-train" {
		t.Errorf("metadata.name: got %v want %q", meta["name"], "ml-train")
	}
	if meta["namespace"] != "tenant-acme" {
		t.Errorf("metadata.namespace: got %v want %q", meta["namespace"], "tenant-acme")
	}

	s, _ := obj["spec"].(map[string]interface{})
	if s["site"] != "paris-dc1" {
		t.Errorf("spec.site: got %v want %q", s["site"], "paris-dc1")
	}
	if s["machineClass"] != "gpu-large" {
		t.Errorf("spec.machineClass: got %v want %q", s["machineClass"], "gpu-large")
	}
	if s["machineCount"] != int64(3) {
		t.Errorf("spec.machineCount: got %v want 3", s["machineCount"])
	}
	if s["addonProfile"] != "gpu-compute" {
		t.Errorf("spec.addonProfile: got %v want %q", s["addonProfile"], "gpu-compute")
	}
}
