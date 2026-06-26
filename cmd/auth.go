// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/smeltry-io/smeltry/internal/auth"
	"github.com/smeltry-io/smeltry/internal/oidc"
)

// EnvToken is the environment variable used to pass a Bearer token in CI
// without saving anything to disk. It takes precedence over the stored token.
// Usage: SMELTRY_TOKEN=<bearer> smeltry cluster list -n tenant-acme
const EnvToken = "SMELTRY_TOKEN"

// minDeviceExpiry is used when the server omits expires_in in the device auth response.
const minDeviceExpiry = 300 * time.Second

func loginScopes() []string {
	return []string{"openid", "email", "groups", "offline_access"}
}

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with the Smeltry platform",
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthLogoutCmd(), newAuthStatusCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var issuerURL, clientID string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate via OIDC device flow",
		Long: `Authenticate with the Smeltry platform using the OIDC device flow.

For CI/non-interactive environments, skip this command and set the
SMELTRY_TOKEN environment variable instead — the token is never written
to disk and is read fresh on every invocation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientIDChanged := cmd.Flags().Changed("client-id")
			return loginDeviceFlow(cmd, issuerURL, clientID, clientIDChanged)
		},
	}
	cmd.Flags().StringVar(&issuerURL, "issuer-url", "",
		"OIDC issuer URL (saved for future logins, e.g. https://auth.example.com/application/o/smeltry/)")
	cmd.Flags().StringVar(&clientID, "client-id", "smeltry-cli", "OIDC client ID")
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

// loginDeviceFlow performs the OIDC device authorization grant.
// clientIDChanged is true when --client-id was explicitly set by the user.
func loginDeviceFlow(cmd *cobra.Command, issuerURL, clientID string, clientIDChanged bool) error {
	// Resolve issuer URL: flag → saved config → error.
	// Only load saved clientID when --client-id was not explicitly passed.
	if issuerURL == "" {
		cfg, err := auth.LoadConfig()
		if err != nil {
			return fmt.Errorf("--issuer-url is required on first login: %w", err)
		}
		issuerURL = cfg.IssuerURL
		if !clientIDChanged && cfg.ClientID != "" {
			clientID = cfg.ClientID
		}
	}

	ctx := context.Background()
	c := oidc.New()

	doc, err := c.Discover(ctx, issuerURL)
	if err != nil {
		return fmt.Errorf("OIDC discovery: %w", err)
	}

	dar, err := c.StartDeviceAuth(ctx, doc.DeviceAuthEndpoint, clientID, loginScopes())
	if err != nil {
		return fmt.Errorf("starting device flow: %w", err)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "\nOpen this URL in your browser:\n  %s\n\n", dar.VerificationURI)
	fmt.Fprintf(out, "Enter code: %s\n\nWaiting for authentication...\n", dar.UserCode)

	interval := time.Duration(dar.Interval) * time.Second
	deadline := time.Duration(dar.ExpiresIn) * time.Second
	if deadline <= 0 {
		deadline = minDeviceExpiry
	}
	pollCtx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	tr, err := c.PollToken(pollCtx, doc.TokenEndpoint, clientID, dar.DeviceCode, interval)
	if err != nil {
		return fmt.Errorf("device flow: %w", err)
	}

	// Extract claims from the ID token without verifying the signature —
	// the kube-apiserver validates the token on every API call.
	email, groups, expiry, err := extractIDTokenClaims(tr.IDToken, tr.ExpiresIn)
	if err != nil {
		return fmt.Errorf("parsing ID token: %w", err)
	}

	if err := auth.Save(&auth.TokenData{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		Expiry:       expiry,
		Email:        email,
		Groups:       groups,
	}); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	// Persist issuer and client ID for future logins (best-effort).
	_ = auth.SaveConfig(&auth.Config{IssuerURL: issuerURL, ClientID: clientID})

	fmt.Fprintf(out, "\nLogged in as %s\n", email)
	return nil
}

// extractIDTokenClaims decodes the JWT payload without verifying the signature,
// issuer, or audience — it is used only to populate the local token cache
// (email, groups, expiry for display). The kube-apiserver performs full
// cryptographic validation on every API call, so a forged local cache cannot
// grant any Kubernetes access.
func extractIDTokenClaims(idToken string, expiresIn int) (email string, groups []string, expiry time.Time, err error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return "", nil, time.Time{}, fmt.Errorf("malformed ID token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, time.Time{}, fmt.Errorf("decoding ID token payload: %w", err)
	}
	var claims struct {
		Email  string   `json:"email"`
		Groups []string `json:"groups"`
		Exp    int64    `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", nil, time.Time{}, fmt.Errorf("parsing ID token claims: %w", err)
	}
	if claims.Exp > 0 {
		expiry = time.Unix(claims.Exp, 0)
	} else {
		expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}
	return claims.Email, claims.Groups, expiry, nil
}
