// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/smeltry-io/smeltry/internal/addonprofile"
	"github.com/smeltry-io/smeltry/internal/netbox"
	"github.com/smeltry-io/smeltry/internal/siteconfig"
)

func TestPrintSites_Table(t *testing.T) {
	sites := []siteconfig.SiteConfig{
		{Name: "paris-dc1"},
		{Name: "london-dc1"},
	}
	cmd := newAdminSiteCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := printSites(cmd, sites, ""); err != nil {
		t.Fatalf("printSites: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "paris-dc1") {
		t.Errorf("expected paris-dc1 in output:\n%s", out)
	}
	if !strings.Contains(out, "london-dc1") {
		t.Errorf("expected london-dc1 in output:\n%s", out)
	}
}

func TestPrintAddonProfiles_Table(t *testing.T) {
	profiles := []addonprofile.AddonProfile{
		{Name: "gpu-compute", Description: "GPU stack"},
		{Name: "base", Description: ""},
	}
	cmd := newAdminAddonProfileCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := printAddonProfiles(cmd, profiles, ""); err != nil {
		t.Fatalf("printAddonProfiles: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "gpu-compute") {
		t.Errorf("expected gpu-compute in output:\n%s", out)
	}
	if !strings.Contains(out, "GPU stack") {
		t.Errorf("expected description in output:\n%s", out)
	}
}

func TestPrintMachines_Table(t *testing.T) {
	devices := []netbox.Device{
		{
			ID:         1,
			Name:       "server-01",
			Status:     netbox.DeviceStatus{Label: "Active"},
			Site:       netbox.NestedSite{Slug: "paris-dc1"},
			DeviceType: netbox.NestedType{Model: "Dell R740"},
		},
	}
	cmd := newAdminMachineCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := printMachines(cmd, devices, ""); err != nil {
		t.Fatalf("printMachines: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "server-01") {
		t.Errorf("expected server-01 in output:\n%s", out)
	}
	if !strings.Contains(out, "paris-dc1") {
		t.Errorf("expected site in output:\n%s", out)
	}
}

func TestPrintMachines_JSON(t *testing.T) {
	devices := []netbox.Device{{ID: 42, Name: "srv"}}
	cmd := newAdminMachineCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := printMachines(cmd, devices, "json"); err != nil {
		t.Fatalf("printMachines json: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "srv") {
		t.Errorf("expected device name in JSON output, got:\n%s", out)
	}
	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Errorf("expected JSON array, got:\n%s", out)
	}
}
