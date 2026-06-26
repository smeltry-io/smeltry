// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"github.com/spf13/cobra"
)

func newAuditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "View audit events",
	}
	cmd.AddCommand(newAuditListCmd())
	return cmd
}

func newAuditListCmd() *cobra.Command {
	var eventType, since string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List AuditEvents in the given namespace",
		RunE:  notImplemented,
	}
	cmd.Flags().StringVar(&eventType, "type", "", "Filter by event type (e.g. PhaseChanged, MachineAllocated)")
	cmd.Flags().StringVar(&since, "since", "", "Show events newer than this duration (e.g. 24h)")
	return cmd
}
