// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var (
		version      string
		installNS    string
		values       []string
		setValues    []string
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install smeltry-operator on the management cluster via Helm",
		RunE:  notImplemented,
	}
	cmd.Flags().StringVar(&version, "version", "latest", "Chart version to install")
	// install uses its own --install-namespace to avoid shadowing the global -n flag.
	cmd.Flags().StringVar(&installNS, "install-namespace", "smeltry-system", "Namespace to deploy the operator into")
	cmd.Flags().StringArrayVarP(&values, "values", "f", nil, "Values file(s) (repeatable)")
	cmd.Flags().StringArrayVar(&setValues, "set", nil, "Set individual values (repeatable, e.g. --set key=val)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print manifests without applying")
	return cmd
}
