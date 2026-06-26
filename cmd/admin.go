// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"github.com/spf13/cobra"
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

func newAdminMachineCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "machine", Short: "Manage physical machines (source: Netbox)"}
	cmd.AddCommand(&cobra.Command{Use: "list", Short: "List machines from Netbox", RunE: notImplemented})
	return cmd
}

func newAdminSiteCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "site", Short: "Manage SiteConfigs"}
	cmd.AddCommand(&cobra.Command{Use: "list", Short: "List SiteConfigs", RunE: notImplemented})
	return cmd
}

func newAdminAddonProfileCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "addon-profile", Short: "Manage AddonProfiles"}
	cmd.AddCommand(&cobra.Command{Use: "list", Short: "List AddonProfiles", RunE: notImplemented})
	return cmd
}

func newAdminTenantCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "tenant", Short: "Manage tenants (namespaces + quotas)"}
	cmd.AddCommand(&cobra.Command{Use: "list", Short: "List tenant namespaces and their quotas", RunE: notImplemented})
	return cmd
}
