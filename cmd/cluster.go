// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/smeltry-io/smeltry/internal/clusterclaim"
	"github.com/smeltry-io/smeltry/internal/k8sclient"
	"github.com/smeltry-io/smeltry/internal/table"
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
		PreRunE: requireNamespace,
		RunE: func(cmd *cobra.Command, args []string) error {
			dyn, err := k8sclient.New(global.Server)
			if err != nil {
				return err
			}
			items, err := clusterclaim.NewClient(dyn).List(context.Background(), global.Namespace)
			if err != nil {
				return err
			}
			return printClusters(cmd, items, global.Output)
		},
	}
}

func newClusterGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get <name>",
		Short:   "Show details of a ClusterClaim",
		Args:    cobra.ExactArgs(1),
		PreRunE: requireNamespace,
		RunE: func(cmd *cobra.Command, args []string) error {
			dyn, err := k8sclient.New(global.Server)
			if err != nil {
				return err
			}
			cc, err := clusterclaim.NewClient(dyn).Get(context.Background(), global.Namespace, args[0])
			if err != nil {
				return err
			}
			return printCluster(cmd, *cc, global.Output)
		},
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

// ── output helpers ────────────────────────────────────────────────────────────

func printClusters(cmd *cobra.Command, items []clusterclaim.ClusterClaim, output string) error {
	switch output {
	case "json":
		return encodeJSON(cmd, items)
	case "yaml":
		return encodeYAML(cmd, items)
	default:
		t := table.New(cmd.OutOrStdout())
		t.Header("NAME", "NAMESPACE", "PHASE", "SITE", "CLASS", "NODES", "AGE")
		for _, cc := range items {
			if err := t.Append(cc.Name, cc.Namespace, cc.Phase, cc.Site, cc.Class,
				fmt.Sprintf("%d", cc.Count), cc.Age); err != nil {
				return err
			}
		}
		return t.Render()
	}
}

func printCluster(cmd *cobra.Command, cc clusterclaim.ClusterClaim, output string) error {
	switch output {
	case "json":
		return encodeJSON(cmd, cc)
	case "yaml":
		return encodeYAML(cmd, cc)
	default:
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Name:       %s\n", cc.Name)
		fmt.Fprintf(out, "Namespace:  %s\n", cc.Namespace)
		fmt.Fprintf(out, "Phase:      %s\n", cc.Phase)
		fmt.Fprintf(out, "Site:       %s\n", cc.Site)
		fmt.Fprintf(out, "Class:      %s\n", cc.Class)
		fmt.Fprintf(out, "Nodes:      %d\n", cc.Count)
		if cc.Age != "" {
			fmt.Fprintf(out, "Age:        %s\n", cc.Age)
		}
		return nil
	}
}

func encodeJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func encodeYAML(cmd *cobra.Command, v any) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(), string(b))
	return err
}

// requireNamespace is a PreRunE that enforces the --namespace flag.
func requireNamespace(cmd *cobra.Command, _ []string) error {
	if global.Namespace == "" {
		return fmt.Errorf("--namespace / -n is required")
	}
	return nil
}

// humanAge returns a human-readable duration since t.
func humanAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
