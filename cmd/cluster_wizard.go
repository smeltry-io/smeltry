// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/smeltry-io/smeltry/internal/addonprofile"
	"github.com/smeltry-io/smeltry/internal/siteconfig"
)

// errWizardAborted is returned when the user declines the final confirmation.
var errWizardAborted = errors.New("wizard aborted by user")

// clusterSpec holds the values collected by the wizard.
type clusterSpec struct {
	Name         string
	Site         string
	MachineClass string
	MachineCount int64
	AddonProfile string
}

// runWizard drives the interactive cluster-creation wizard.
// w receives all prompts; r provides user input; sites and profiles are
// pre-fetched from the cluster so the wizard itself needs no network access.
func runWizard(w io.Writer, r io.Reader, sites []siteconfig.SiteConfig, profiles []addonprofile.AddonProfile) (clusterSpec, error) {
	if len(sites) == 0 {
		return clusterSpec{}, errors.New("no SiteConfigs available — ask an admin to create one")
	}
	if len(profiles) == 0 {
		return clusterSpec{}, errors.New("no AddonProfiles available — ask an admin to create one")
	}

	sc := bufio.NewScanner(r)
	prompt := func(q string) string {
		fmt.Fprint(w, q)
		sc.Scan()
		return strings.TrimSpace(sc.Text())
	}

	var spec clusterSpec

	// Name
	spec.Name = prompt("Cluster name: ")

	// Site
	fmt.Fprintln(w, "\nAvailable sites:")
	for i, s := range sites {
		fmt.Fprintf(w, "  %d. %s\n", i+1, s.Name)
	}
	siteRaw := prompt(fmt.Sprintf("Site [1-%d] (default 1): ", len(sites)))
	spec.Site = pickName(siteRaw, func(i int) string { return sites[i].Name }, len(sites))

	// Machine class
	spec.MachineClass = prompt("\nMachine class: ")

	// Machine count
	countRaw := prompt("Machine count [1]: ")
	if countRaw == "" {
		spec.MachineCount = 1
	} else if n, err := strconv.ParseInt(countRaw, 10, 64); err == nil && n > 0 {
		spec.MachineCount = n
	} else {
		spec.MachineCount = 1
	}

	// Addon profile
	fmt.Fprintln(w, "\nAvailable addon profiles:")
	for i, p := range profiles {
		if p.Description != "" {
			fmt.Fprintf(w, "  %d. %s — %s\n", i+1, p.Name, p.Description)
		} else {
			fmt.Fprintf(w, "  %d. %s\n", i+1, p.Name)
		}
	}
	profileRaw := prompt(fmt.Sprintf("Addon profile [1-%d] (default 1): ", len(profiles)))
	spec.AddonProfile = pickName(profileRaw, func(i int) string { return profiles[i].Name }, len(profiles))

	// Summary + confirm
	fmt.Fprintf(w, "\nCreating ClusterClaim %q:\n", spec.Name)
	fmt.Fprintf(w, "  Site:          %s\n", spec.Site)
	fmt.Fprintf(w, "  Machine class: %s\n", spec.MachineClass)
	fmt.Fprintf(w, "  Machine count: %d\n", spec.MachineCount)
	fmt.Fprintf(w, "  Addon profile: %s\n", spec.AddonProfile)
	confirm := prompt("\nConfirm? [y/N]: ")
	if !strings.EqualFold(confirm, "y") && !strings.EqualFold(confirm, "yes") {
		return clusterSpec{}, errWizardAborted
	}

	return spec, nil
}

// pickName converts a 1-based numeric choice string into the corresponding
// name. Out-of-range or non-numeric input falls back to index 0.
func pickName(raw string, name func(int) string, count int) string {
	if n, err := strconv.Atoi(raw); err == nil && n >= 1 && n <= count {
		return name(n - 1)
	}
	return name(0)
}

// clusterSpecToObject builds the unstructured map passed to the dynamic client.
func clusterSpecToObject(namespace string, spec clusterSpec) map[string]interface{} {
	obj := map[string]interface{}{
		"apiVersion": "portal.smeltry.io/v1alpha1",
		"kind":       "ClusterClaim",
		"metadata": map[string]interface{}{
			"name":      spec.Name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"site":         spec.Site,
			"machineClass": spec.MachineClass,
			"machineCount": spec.MachineCount,
			"addonProfile": spec.AddonProfile,
		},
	}
	return obj
}
