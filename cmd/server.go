// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"github.com/spf13/cobra"
)

func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Manage bare-metal server claims",
	}
	cmd.AddCommand(
		newServerListCmd(),
		newServerGetCmd(),
		newServerCreateCmd(),
		newServerDeleteCmd(),
	)
	return cmd
}

func newServerListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List ServerClaims in the given namespace",
		RunE:  notImplemented,
	}
}

func newServerGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Show details of a ServerClaim",
		Args:  cobra.ExactArgs(1),
		RunE:  notImplemented,
	}
}

func newServerCreateCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a ServerClaim (interactive wizard or --file)",
		RunE:  notImplemented,
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to a ServerClaim manifest (skips interactive wizard)")
	return cmd
}

func newServerDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a ServerClaim",
		Args:  cobra.ExactArgs(1),
		RunE:  notImplemented,
	}
}
