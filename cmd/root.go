// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"github.com/spf13/cobra"
)

// GlobalFlags holds flags shared across all commands.
type GlobalFlags struct {
	Namespace  string
	Server     string
	Output     string
	Insecure   bool
	Wait       bool
	Timeout    string
}

var global GlobalFlags

var rootCmd = &cobra.Command{
	Use:   "smeltry",
	Short: "Self-service bare metal infrastructure CLI",
	Long: `smeltry is the command-line interface for the Smeltry platform.
It lets tenants claim Kubernetes clusters and bare-metal servers,
and lets admins inspect machines, tenants, and operator health.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&global.Namespace, "namespace", "n", "", "Target namespace (required for most commands)")
	rootCmd.PersistentFlags().StringVar(&global.Server, "server", "", "Override kube-apiserver URL")
	rootCmd.PersistentFlags().StringVarP(&global.Output, "output", "o", "table", "Output format: table|json|yaml|wide")
	rootCmd.PersistentFlags().BoolVar(&global.Insecure, "insecure-skip-tls-verify", false, "Skip TLS certificate verification")
	rootCmd.PersistentFlags().BoolVar(&global.Wait, "wait", false, "Wait until the operation reaches the desired state")
	rootCmd.PersistentFlags().StringVar(&global.Timeout, "timeout", "", "Maximum wait duration (e.g. 5m, 30s) — only used with --wait")

	rootCmd.AddCommand(
		newAuthCmd(),
		newClusterCmd(),
		newServerCmd(),
		newAuditCmd(),
		newInstallCmd(),
		newStatusCmd(),
		newAdminCmd(),
	)
}
