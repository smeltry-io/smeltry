// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"github.com/spf13/cobra"
)

func newClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Kubernetes cluster claims",
	}
	cmd.AddCommand(
		newClusterListCmd(),
		newClusterGetCmd(),
		newClusterCreateCmd(),
		newClusterDeleteCmd(),
		newClusterKubeconfigCmd(),
		newClusterAddonsCmd(),
	)
	return cmd
}

func newClusterListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List ClusterClaims in the given namespace",
		RunE:  notImplemented,
	}
}

func newClusterGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Show details of a ClusterClaim",
		Args:  cobra.ExactArgs(1),
		RunE:  notImplemented,
	}
}

func newClusterCreateCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a ClusterClaim (interactive wizard or --file)",
		RunE:  notImplemented,
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to a ClusterClaim manifest (skips interactive wizard)")
	return cmd
}

func newClusterDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a ClusterClaim",
		Args:  cobra.ExactArgs(1),
		RunE:  notImplemented,
	}
}

func newClusterKubeconfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kubeconfig <name>",
		Short: "Print the Headlamp deep-link to download the cluster kubeconfig",
		Args:  cobra.ExactArgs(1),
		RunE:  notImplemented,
	}
}

func newClusterAddonsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "addons <name>",
		Short: "List addon HelmReleases and their status for a ClusterClaim",
		Args:  cobra.ExactArgs(1),
		RunE:  notImplemented,
	}
}
