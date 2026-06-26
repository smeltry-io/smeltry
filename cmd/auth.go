// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/smeltry-io/smeltry/internal/auth"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with the Smeltry platform",
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthLogoutCmd(), newAuthStatusCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var tokenFlag string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate via OIDC device flow (or --token for CI)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenFlag != "" {
				return loginWithToken(tokenFlag)
			}
			return loginDeviceFlow()
		},
	}
	cmd.Flags().StringVar(&tokenFlag, "token", "", "Use a pre-obtained Bearer token (for CI/non-interactive use)")
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

// loginWithToken stores a pre-obtained access token directly.
// Groups and email are not decoded from the token in v1 — the user provides
// only the raw Bearer string (suitable for CI pipelines).
func loginWithToken(token string) error {
	return auth.Save(&auth.TokenData{
		AccessToken: token,
		// Expiry left zero — IsExpired() will return true immediately on the
		// next command, forcing a re-login. For CI each invocation provides --token.
	})
}

// loginDeviceFlow is a placeholder for the OIDC device flow implementation.
// It will be wired to the Authentik issuer in a subsequent story.
func loginDeviceFlow() error {
	return fmt.Errorf("device flow not yet implemented — use --token for now")
}
