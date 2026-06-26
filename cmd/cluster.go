// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/smeltry-io/smeltry/internal/addon"
	"github.com/smeltry-io/smeltry/internal/addonprofile"
	"github.com/smeltry-io/smeltry/internal/clusterclaim"
	"github.com/smeltry-io/smeltry/internal/k8sclient"
	"github.com/smeltry-io/smeltry/internal/poller"
	"github.com/smeltry-io/smeltry/internal/siteconfig"
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
		Use:     "create",
		Short:   "Create a ClusterClaim (interactive wizard or --file)",
		PreRunE: requireNamespace,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file != "" {
				return clusterCreateFromFile(cmd, file)
			}
			return clusterCreateWizard(cmd)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to a ClusterClaim manifest (skips interactive wizard)")
	return cmd
}

func newClusterDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Short:   "Delete a ClusterClaim",
		Args:    cobra.ExactArgs(1),
		PreRunE: requireNamespace,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Interactive confirmation.
			if !confirmResourceDelete(cmd.OutOrStdout(), os.Stdin, "ClusterClaim", name, global.Namespace) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}

			dyn, err := k8sclient.New(global.Server)
			if err != nil {
				return err
			}
			cc := clusterclaim.NewClient(dyn)

			if err := cc.Delete(context.Background(), global.Namespace, name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ClusterClaim %q deleted.\n", name)

			if !global.Wait {
				return nil
			}
			timeout, err := waitTimeout()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			fmt.Fprintf(cmd.OutOrStdout(), "Waiting for %q to be fully removed...\n", name)
			return poller.UntilDone(ctx, poller.DefaultInterval, func(ctx context.Context) (bool, error) {
				return cc.IsGone(ctx, global.Namespace, name)
			})
		},
	}
}

func newClusterKubeconfigCmd() *cobra.Command {
	var headlampURL string
	cmd := &cobra.Command{
		Use:     "kubeconfig <name>",
		Short:   "Print the Headlamp deep-link to download the cluster kubeconfig",
		Args:    cobra.ExactArgs(1),
		PreRunE: requireNamespace,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			// The kubeconfig is stored in secret "<name>-kubeconfig" in the tenant
			// namespace. Tenants cannot list secrets globally, so the download goes
			// through Headlamp which has the necessary RBAC. We print a deep-link.
			secretName := fmt.Sprintf("%s-kubeconfig", name)
			link := fmt.Sprintf("%s/c/local/namespaces/%s/secrets/%s",
				strings.TrimRight(headlampURL, "/"), global.Namespace, secretName)
			fmt.Fprintf(cmd.OutOrStdout(),
				"Download kubeconfig for %q via Headlamp:\n  %s\n", name, link)
			return nil
		},
	}
	cmd.Flags().StringVar(&headlampURL, "headlamp-url", "http://localhost:4466",
		"Base URL of the Headlamp instance")
	return cmd
}

func newClusterAddonsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "addons <name>",
		Short:   "List addon HelmReleases and their status for a ClusterClaim",
		Args:    cobra.ExactArgs(1),
		PreRunE: requireNamespace,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			dyn, err := k8sclient.New(global.Server)
			if err != nil {
				return err
			}
			items, err := addon.NewClient(dyn).ListForCluster(context.Background(), global.Namespace, name)
			if err != nil {
				return err
			}
			t := table.New(cmd.OutOrStdout())
			t.Header("NAME", "READY", "BOOTSTRAP")
			for _, hr := range items {
				if err := t.Append(hr.Name, boolStr(hr.Ready), boolStr(hr.Bootstrap)); err != nil {
					return err
				}
			}
			return t.Render()
		},
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

// clusterCreateWizard runs the interactive wizard and creates the ClusterClaim.
func clusterCreateWizard(cmd *cobra.Command) error {
	dyn, err := k8sclient.New(global.Server)
	if err != nil {
		return err
	}
	sites, err := siteconfig.NewClient(dyn).List(context.Background())
	if err != nil {
		return err
	}
	profiles, err := addonprofile.NewClient(dyn).List(context.Background())
	if err != nil {
		return err
	}
	spec, err := runWizard(cmd.OutOrStdout(), os.Stdin, sites, profiles)
	if errors.Is(err, errWizardAborted) {
		fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
		return nil
	}
	if err != nil {
		return err
	}
	obj := clusterSpecToObject(global.Namespace, spec)
	cc, err := clusterclaim.NewClient(dyn).Create(context.Background(), global.Namespace, obj)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "ClusterClaim %q created (phase: %s).\n", cc.Name, cc.Phase)
	return nil
}

// clusterCreateFromFile reads a YAML/JSON manifest and creates the ClusterClaim.
func clusterCreateFromFile(cmd *cobra.Command, path string) error {
	obj, ns, err := loadManifest(path)
	if err != nil {
		return err
	}
	if ns != "" && ns != global.Namespace {
		fmt.Fprintf(cmd.OutOrStdout(),
			"Warning: manifest namespace %q differs from --namespace %q; using %q\n",
			ns, global.Namespace, global.Namespace)
	}
	dyn, err := k8sclient.New(global.Server)
	if err != nil {
		return err
	}
	cc, err := clusterclaim.NewClient(dyn).Create(context.Background(), global.Namespace, obj)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "ClusterClaim %q created (phase: %s).\n", cc.Name, cc.Phase)
	return nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// waitTimeout parses global.Timeout or returns a sensible default (10m).
func waitTimeout() (time.Duration, error) {
	if global.Timeout == "" {
		return 10 * time.Minute, nil
	}
	d, err := time.ParseDuration(global.Timeout)
	if err != nil {
		return 0, fmt.Errorf("invalid --timeout %q: %w", global.Timeout, err)
	}
	return d, nil
}

// requireNamespace is a PreRunE that enforces the --namespace flag.
func requireNamespace(cmd *cobra.Command, _ []string) error {
	if global.Namespace == "" {
		return fmt.Errorf("--namespace / -n is required")
	}
	return nil
}

