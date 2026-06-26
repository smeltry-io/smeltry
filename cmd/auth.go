// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/smeltry-io/smeltry/internal/auth"
)

// EnvToken is the environment variable used to pass a Bearer token in CI
// without saving anything to disk. It takes precedence over the stored token.
// Usage: SMELTRY_TOKEN=<bearer> smeltry cluster list -n tenant-acme
const EnvToken = "SMELTRY_TOKEN"

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with the Smeltry platform",
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthLogoutCmd(), newAuthStatusCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate via OIDC device flow",
		Long: `Authenticate with the Smeltry platform using the OIDC device flow.

For CI/non-interactive environments, skip this command and set the
SMELTRY_TOKEN environment variable instead — the token is never written
to disk and is read fresh on every invocation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return loginDeviceFlow()
		},
	}
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.Delete(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if t := os.Getenv(EnvToken); t != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Mode:    CI (SMELTRY_TOKEN env var)")
				fmt.Fprintln(cmd.OutOrStdout(), "Status:  token provided via environment, not stored on disk")
				return nil
			}
			td, err := auth.Load()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Email:   %s\n", td.Email)
			fmt.Fprintf(out, "Groups:  %v\n", td.Groups)
			fmt.Fprintf(out, "Expiry:  %s\n", td.Expiry.Format("2006-01-02 15:04:05 MST"))
			if td.IsExpired() {
				fmt.Fprintln(out, "Status:  EXPIRED")
			} else {
				fmt.Fprintln(out, "Status:  valid")
			}
			return nil
		},
	}
}

// loginDeviceFlow is a placeholder for the OIDC device flow implementation.
// It will be wired to the Authentik issuer in a subsequent story.
func loginDeviceFlow() error {
	return fmt.Errorf("device flow not yet implemented — set SMELTRY_TOKEN for now")
}
