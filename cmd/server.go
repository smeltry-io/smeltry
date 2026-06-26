// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/smeltry-io/smeltry/internal/k8sclient"
	"github.com/smeltry-io/smeltry/internal/poller"
	"github.com/smeltry-io/smeltry/internal/serverclaim"
	"github.com/smeltry-io/smeltry/internal/siteconfig"
	"github.com/smeltry-io/smeltry/internal/table"
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
		Use:     "list",
		Short:   "List ServerClaims in the given namespace",
		PreRunE: requireNamespace,
		RunE: func(cmd *cobra.Command, args []string) error {
			dyn, err := k8sclient.New(global.Server)
			if err != nil {
				return err
			}
			items, err := serverclaim.NewClient(dyn).List(context.Background(), global.Namespace)
			if err != nil {
				return err
			}
			return printServers(cmd, items, global.Output)
		},
	}
}

func newServerGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get <name>",
		Short:   "Show details of a ServerClaim",
		Args:    cobra.ExactArgs(1),
		PreRunE: requireNamespace,
		RunE: func(cmd *cobra.Command, args []string) error {
			dyn, err := k8sclient.New(global.Server)
			if err != nil {
				return err
			}
			sc, err := serverclaim.NewClient(dyn).Get(context.Background(), global.Namespace, args[0])
			if err != nil {
				return err
			}
			return printServer(cmd, *sc, global.Output)
		},
	}
}

func newServerCreateCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a ServerClaim (interactive wizard or --file)",
		PreRunE: requireNamespace,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file != "" {
				return serverCreateFromFile(cmd, file)
			}
			return serverCreateWizard(cmd)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to a ServerClaim manifest (skips interactive wizard)")
	return cmd
}

func newServerDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Short:   "Delete a ServerClaim",
		Args:    cobra.ExactArgs(1),
		PreRunE: requireNamespace,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !confirmResourceDelete(cmd.OutOrStdout(), os.Stdin, "ServerClaim", name, global.Namespace) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}

			dyn, err := k8sclient.New(global.Server)
			if err != nil {
				return err
			}
			sc := serverclaim.NewClient(dyn)

			if err := sc.Delete(context.Background(), global.Namespace, name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ServerClaim %q deleted.\n", name)

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
				return sc.IsGone(ctx, global.Namespace, name)
			})
		},
	}
}

// ── output helpers ────────────────────────────────────────────────────────────

func printServers(cmd *cobra.Command, items []serverclaim.ServerClaim, output string) error {
	switch output {
	case "json":
		return encodeJSON(cmd, items)
	case "yaml":
		return encodeYAML(cmd, items)
	default:
		t := table.New(cmd.OutOrStdout())
		t.Header("NAME", "NAMESPACE", "PHASE", "SITE", "CLASS", "OS", "IP", "AGE")
		for _, sc := range items {
			if err := t.Append(sc.Name, sc.Namespace, sc.Phase, sc.Site, sc.Class,
				sc.OS, sc.ServerIP, sc.Age); err != nil {
				return err
			}
		}
		return t.Render()
	}
}

func printServer(cmd *cobra.Command, sc serverclaim.ServerClaim, output string) error {
	switch output {
	case "json":
		return encodeJSON(cmd, sc)
	case "yaml":
		return encodeYAML(cmd, sc)
	default:
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Name:       %s\n", sc.Name)
		fmt.Fprintf(out, "Namespace:  %s\n", sc.Namespace)
		fmt.Fprintf(out, "Phase:      %s\n", sc.Phase)
		fmt.Fprintf(out, "Site:       %s\n", sc.Site)
		fmt.Fprintf(out, "Class:      %s\n", sc.Class)
		fmt.Fprintf(out, "OS:         %s\n", sc.OS)
		if sc.ServerIP != "" {
			fmt.Fprintf(out, "IP:         %s\n", sc.ServerIP)
		}
		if sc.Age != "" {
			fmt.Fprintf(out, "Age:        %s\n", sc.Age)
		}
		return nil
	}
}

// serverCreateWizard runs the interactive wizard and creates the ServerClaim.
func serverCreateWizard(cmd *cobra.Command) error {
	dyn, err := k8sclient.New(global.Server)
	if err != nil {
		return err
	}
	sites, err := siteconfig.NewClient(dyn).List(context.Background())
	if err != nil {
		return err
	}
	spec, err := runServerWizard(cmd.OutOrStdout(), os.Stdin, sites)
	if errors.Is(err, errWizardAborted) {
		fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
		return nil
	}
	if err != nil {
		return err
	}
	obj := serverSpecToObject(global.Namespace, spec)
	sc, err := serverclaim.NewClient(dyn).Create(context.Background(), global.Namespace, obj)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "ServerClaim %q created (phase: %s).\n", sc.Name, sc.Phase)
	return nil
}

// serverCreateFromFile reads a YAML/JSON manifest and creates the ServerClaim.
func serverCreateFromFile(cmd *cobra.Command, path string) error {
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
	sc, err := serverclaim.NewClient(dyn).Create(context.Background(), global.Namespace, obj)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "ServerClaim %q created (phase: %s).\n", sc.Name, sc.Phase)
	return nil
}

// confirmResourceDelete prompts the user to confirm deletion by typing the
// resource name. kind is the CRD kind name for the prompt message.
func confirmResourceDelete(w io.Writer, r io.Reader, kind, name, namespace string) bool {
	fmt.Fprintf(w, "Delete %s %q from namespace %q? Type the name to confirm: ", kind, name, namespace)
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text()) == name
}
