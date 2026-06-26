// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/smeltry-io/smeltry/internal/helminstall"
)

func newInstallCmd() *cobra.Command {
	var (
		version   string
		installNS string
		values    []string
		setValues []string
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install smeltry-operator on the management cluster via Helm",
		Long: `Install or upgrade the smeltry-operator Helm chart on the management cluster.

On first run this creates the target namespace and installs the chart.
On subsequent runs it upgrades the existing release in-place.

The chart is pulled from the OCI registry:
  ` + helminstall.DefaultChart,
		RunE: func(cmd *cobra.Command, args []string) error {
			timeout, err := waitTimeout()
			if err != nil {
				return err
			}
			if !global.Wait {
				timeout = 5 * time.Minute // sensible default even without --wait
			}

			opts := helminstall.Options{
				Namespace:  installNS,
				Version:    version,
				ValueFiles: values,
				SetValues:  setValues,
				DryRun:     dryRun,
				Wait:       global.Wait,
				Timeout:    timeout,
			}

			if dryRun {
				fmt.Fprintln(cmd.OutOrStdout(), "(dry-run mode — no changes will be applied)")
			}

			return helminstall.Install(cmd.OutOrStdout(), opts)
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Chart version to install (default: latest)")
	// install uses its own --install-namespace to avoid shadowing the global -n flag.
	cmd.Flags().StringVar(&installNS, "install-namespace", "smeltry-system",
		"Namespace to deploy the operator into")
	cmd.Flags().StringArrayVarP(&values, "values", "f", nil,
		"Values file(s) (repeatable)")
	cmd.Flags().StringArrayVar(&setValues, "set", nil,
		"Set individual values (repeatable, e.g. --set key=val)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Print manifests without applying")
	return cmd
}
