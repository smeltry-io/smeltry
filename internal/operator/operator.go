// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package operator fetches the runtime status of the smeltry-operator
// deployment and the counts of active CRD resources.
package operator

import (
	"context"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	operatorName      = "smeltry-operator"
	operatorNamespace = "smeltry-system"
)

var (
	deployGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	ccGVR     = schema.GroupVersionResource{Group: "portal.smeltry.io", Version: "v1alpha1", Resource: "clusterclaims"}
	scGVR     = schema.GroupVersionResource{Group: "portal.smeltry.io", Version: "v1alpha1", Resource: "serverclaims"}
)

// Status holds the aggregated state of the smeltry-operator installation.
type Status struct {
	Absent        bool
	Ready         bool
	Replicas      int64
	ReadyReplicas int64
	Version       string
	ClusterClaims int
	ServerClaims  int
}

// Fetch queries the cluster for operator deployment status and resource counts.
func Fetch(ctx context.Context, dyn dynamic.Interface) (Status, error) {
	var st Status

	obj, err := dyn.Resource(deployGVR).Namespace(operatorNamespace).Get(ctx, operatorName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			st.Absent = true
			return st, nil
		}
		return st, fmt.Errorf("getting operator deployment: %w", err)
	}

	labels, _ := obj.Object["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
	st.Version, _ = labels["app.kubernetes.io/version"].(string)

	status, _ := obj.Object["status"].(map[string]interface{})
	st.Replicas = intVal(status["replicas"])
	st.ReadyReplicas = intVal(status["readyReplicas"])
	st.Ready = st.ReadyReplicas > 0 && st.ReadyReplicas == st.Replicas

	cc, err := dyn.Resource(ccGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return st, fmt.Errorf("listing ClusterClaims: %w", err)
	}
	if cc != nil {
		st.ClusterClaims = len(cc.Items)
	}

	sc, err := dyn.Resource(scGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return st, fmt.Errorf("listing ServerClaims: %w", err)
	}
	if sc != nil {
		st.ServerClaims = len(sc.Items)
	}

	return st, nil
}

func intVal(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	}
	return 0
}
