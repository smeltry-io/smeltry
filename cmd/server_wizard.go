// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/smeltry-io/smeltry/internal/siteconfig"
)

func supportedOS() []string { return []string{"flatcar", "ubuntu"} }

type serverSpec struct {
	name         string
	site         string
	machineClass string
	os           string
}

// runServerWizard drives an interactive prompt to build a serverSpec.
func runServerWizard(w io.Writer, r io.Reader, sites []siteconfig.SiteConfig) (serverSpec, error) {
	if len(sites) == 0 {
		return serverSpec{}, fmt.Errorf("no SiteConfigs available — ask an admin to create one")
	}
	scan := bufio.NewScanner(r)

	site, err := pickSite(w, scan, sites)
	if err != nil {
		return serverSpec{}, err
	}

	machineClass, err := promptString(w, scan, "Machine class (e.g. standard, gpu-large)")
	if err != nil {
		return serverSpec{}, err
	}

	os, err := pickOS(w, scan)
	if err != nil {
		return serverSpec{}, err
	}

	name, err := promptString(w, scan, "Name for this server claim")
	if err != nil {
		return serverSpec{}, err
	}

	return serverSpec{name: name, site: site, machineClass: machineClass, os: os}, nil
}

func pickSite(w io.Writer, scan *bufio.Scanner, sites []siteconfig.SiteConfig) (string, error) {
	fmt.Fprintln(w, "\nAvailable sites:")
	for i, s := range sites {
		fmt.Fprintf(w, "  [%d] %s\n", i+1, s.Name)
	}
	fmt.Fprint(w, "Select site (number): ")
	if !scan.Scan() {
		return "", errWizardAborted
	}
	raw := strings.TrimSpace(scan.Text())
	if raw == "q" || raw == "" {
		return "", errWizardAborted
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > len(sites) {
		return "", fmt.Errorf("invalid site selection %q", raw)
	}
	return sites[n-1].Name, nil
}

func pickOS(w io.Writer, scan *bufio.Scanner) (string, error) {
	osList := supportedOS()
	fmt.Fprintln(w, "\nAvailable operating systems:")
	for i, osName := range osList {
		fmt.Fprintf(w, "  [%d] %s\n", i+1, osName)
	}
	fmt.Fprint(w, "Select OS (number): ")
	if !scan.Scan() {
		return "", errWizardAborted
	}
	raw := strings.TrimSpace(scan.Text())
	if raw == "q" || raw == "" {
		return "", errWizardAborted
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > len(osList) {
		return "", fmt.Errorf("invalid OS selection %q", raw)
	}
	return osList[n-1], nil
}

func promptString(w io.Writer, scan *bufio.Scanner, prompt string) (string, error) {
	fmt.Fprintf(w, "%s: ", prompt)
	if !scan.Scan() {
		return "", errWizardAborted
	}
	val := strings.TrimSpace(scan.Text())
	if val == "q" || val == "" {
		return "", errWizardAborted
	}
	return val, nil
}

// serverSpecToObject converts a serverSpec to the unstructured object expected
// by the dynamic client.
func serverSpecToObject(namespace string, spec serverSpec) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "portal.smeltry.io/v1alpha1",
		"kind":       "ServerClaim",
		"metadata": map[string]interface{}{
			"name":      spec.name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"site":         spec.site,
			"machineClass": spec.machineClass,
			"os":           spec.os,
		},
	}
}
