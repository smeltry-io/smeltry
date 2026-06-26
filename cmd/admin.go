// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smeltry-io/smeltry/internal/addonprofile"
	"github.com/smeltry-io/smeltry/internal/k8sclient"
	"github.com/smeltry-io/smeltry/internal/netbox"
	"github.com/smeltry-io/smeltry/internal/siteconfig"
	"github.com/smeltry-io/smeltry/internal/table"
	"github.com/smeltry-io/smeltry/internal/tenant"
)

func newAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Admin commands (requires smeltry-admins group)",
	}
	cmd.AddCommand(
		newAdminMachineCmd(),
		newAdminSiteCmd(),
		newAdminAddonProfileCmd(),
		newAdminTenantCmd(),
	)
	return cmd
}

// ── machine ───────────────────────────────────────────────────────────────────

func newAdminMachineCmd() *cobra.Command {
	var (
		netboxURL   string
		netboxToken string
		site        string
		status      string
	)
	cmd := &cobra.Command{
		Use:   "machine",
		Short: "Manage physical machines (source: Netbox)",
	}
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List machines from Netbox",
		RunE: func(c *cobra.Command, args []string) error {
			token := netboxToken
			if token == "" {
				token = os.Getenv("NETBOX_TOKEN")
			}
			if token == "" {
				return fmt.Errorf("Netbox token required: set --netbox-token or NETBOX_TOKEN")
			}
			if netboxURL == "" {
				return fmt.Errorf("Netbox URL required: set --netbox-url")
			}
			nc := netbox.NewClient(netboxURL, token)
			devices, err := nc.ListDevices(context.Background(), netbox.ListDevicesParams{
				Site:   site,
				Status: status,
			})
			if err != nil {
				return err
			}
			return printMachines(c, devices, global.Output)
		},
	}
	listCmd.Flags().StringVar(&netboxURL, "netbox-url", "", "Netbox base URL (e.g. https://netbox.example.com)")
	listCmd.Flags().StringVar(&netboxToken, "netbox-token", "", "Netbox API token (prefer NETBOX_TOKEN env var to avoid exposure in shell history)")
	listCmd.Flags().Lookup("netbox-token").Hidden = true
	listCmd.Flags().StringVar(&site, "site", "", "Filter by site slug")
	listCmd.Flags().StringVar(&status, "status", "", "Filter by device status (e.g. active, staged)")
	cmd.AddCommand(listCmd)
	return cmd
}

func printMachines(cmd *cobra.Command, devices []netbox.Device, output string) error {
	switch output {
	case "json":
		return encodeJSON(cmd, devices)
	case "yaml":
		return encodeYAML(cmd, devices)
	default:
		t := table.New(cmd.OutOrStdout())
		t.Header("NAME", "STATUS", "SITE", "MODEL")
		for _, d := range devices {
			if err := t.Append(d.Name, d.Status.Label, d.Site.Slug, d.DeviceType.Model); err != nil {
				return err
			}
		}
		return t.Render()
	}
}

// ── site ──────────────────────────────────────────────────────────────────────

func newAdminSiteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "site",
		Short: "Manage SiteConfigs",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List SiteConfigs",
		RunE: func(c *cobra.Command, args []string) error {
			dyn, err := k8sclient.New(global.Server)
			if err != nil {
				return err
			}
			sites, err := siteconfig.NewClient(dyn).List(context.Background())
			if err != nil {
				return err
			}
			return printSites(c, sites, global.Output)
		},
	})
	return cmd
}

func printSites(cmd *cobra.Command, sites []siteconfig.SiteConfig, output string) error {
	switch output {
	case "json":
		return encodeJSON(cmd, sites)
	case "yaml":
		return encodeYAML(cmd, sites)
	default:
		t := table.New(cmd.OutOrStdout())
		t.Header("NAME")
		for _, s := range sites {
			if err := t.Append(s.Name); err != nil {
				return err
			}
		}
		return t.Render()
	}
}

// ── addon-profile ─────────────────────────────────────────────────────────────

func newAdminAddonProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addon-profile",
		Short: "Manage AddonProfiles",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List AddonProfiles",
		RunE: func(c *cobra.Command, args []string) error {
			dyn, err := k8sclient.New(global.Server)
			if err != nil {
				return err
			}
			profiles, err := addonprofile.NewClient(dyn).List(context.Background())
			if err != nil {
				return err
			}
			return printAddonProfiles(c, profiles, global.Output)
		},
	})
	return cmd
}

func printAddonProfiles(cmd *cobra.Command, profiles []addonprofile.AddonProfile, output string) error {
	switch output {
	case "json":
		return encodeJSON(cmd, profiles)
	case "yaml":
		return encodeYAML(cmd, profiles)
	default:
		t := table.New(cmd.OutOrStdout())
		t.Header("NAME", "DESCRIPTION")
		for _, p := range profiles {
			if err := t.Append(p.Name, p.Description); err != nil {
				return err
			}
		}
		return t.Render()
	}
}

// ── tenant ────────────────────────────────────────────────────────────────────

func newAdminTenantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Manage tenants (namespaces + quotas)",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List tenant namespaces and their quotas",
		RunE: func(c *cobra.Command, args []string) error {
			dyn, err := k8sclient.New(global.Server)
			if err != nil {
				return err
			}
			tenants, err := tenant.List(context.Background(), dyn)
			if err != nil {
				return err
			}
			return printTenants(c, tenants, global.Output)
		},
	})
	return cmd
}

func printTenants(cmd *cobra.Command, tenants []tenant.Tenant, output string) error {
	switch output {
	case "json":
		return encodeJSON(cmd, tenants)
	case "yaml":
		return encodeYAML(cmd, tenants)
	default:
		t := table.New(cmd.OutOrStdout())
		t.Header("NAMESPACE", "MAX-CLUSTERS", "MAX-SERVERS")
		for _, tn := range tenants {
			if err := t.Append(tn.Namespace, tn.MaxClusters, tn.MaxNodes); err != nil {
				return err
			}
		}
		return t.Render()
	}
}
