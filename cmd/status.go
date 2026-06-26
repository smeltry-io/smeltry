// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/smeltry-io/smeltry/internal/k8sclient"
	"github.com/smeltry-io/smeltry/internal/operator"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the health of the smeltry-operator and active resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			dyn, err := k8sclient.New(global.Server)
			if err != nil {
				return err
			}
			st, err := operator.Fetch(context.Background(), dyn)
			if err != nil {
				return err
			}
			formatStatus(cmd.OutOrStdout(), st)
			return nil
		},
	}
}

func formatStatus(w io.Writer, st operator.Status) {
	fmt.Fprintln(w, "smeltry-operator")
	fmt.Fprint(w, "  Status:   ")
	switch {
	case st.Absent:
		fmt.Fprintln(w, "not installed (run 'smeltry install')")
	case st.Ready:
		fmt.Fprintln(w, "Ready")
	default:
		fmt.Fprintf(w, "Not Ready (%d/%d replicas ready)\n", st.ReadyReplicas, st.Replicas)
	}
	if st.Version != "" {
		fmt.Fprintf(w, "  Version:  %s\n", st.Version)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Active resources (all namespaces)")
	fmt.Fprintf(w, "  ClusterClaims:  %d\n", st.ClusterClaims)
	fmt.Fprintf(w, "  ServerClaims:   %d\n", st.ServerClaims)
}
