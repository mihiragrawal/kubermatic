/*
Copyright 2020 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"context"
	"fmt"

	kubermaticv1 "k8c.io/kubermatic/v2/pkg/apis/kubermatic/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// EnqueueClusterForNamespacedObject enqueues the cluster that owns a namespaced object, if any
// It is used by various controllers to react to changes in the resources in the cluster namespace.
func EnqueueClusterForNamespacedObject(client ctrlruntimeclient.Client) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(a ctrlruntimeclient.Object) []reconcile.Request {
		clusterList := &kubermaticv1.ClusterList{}
		if err := client.List(context.Background(), clusterList); err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to list Clusters: %w", err))
			return []reconcile.Request{}
		}
		for _, cluster := range clusterList.Items {
			if cluster.Status.NamespaceName == a.GetNamespace() {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: cluster.Name}}}
			}
		}
		return []reconcile.Request{}
	})
}

// EnqueueClusterForNamespacedObjectWithSeedName enqueues the cluster that owns a namespaced object,
// if any. The seedName is put into the namespace field
// It is used by various controllers to react to changes in the resources in the cluster namespace.
func EnqueueClusterForNamespacedObjectWithSeedName(client ctrlruntimeclient.Client, seedName string, clusterSelector labels.Selector) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(a ctrlruntimeclient.Object) []reconcile.Request {
		clusterList := &kubermaticv1.ClusterList{}
		listOpts := &ctrlruntimeclient.ListOptions{
			LabelSelector: clusterSelector,
		}

		if err := client.List(context.Background(), clusterList, listOpts); err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to list Clusters: %w", err))
			return []reconcile.Request{}
		}

		for _, cluster := range clusterList.Items {
			if cluster.Status.NamespaceName == a.GetNamespace() {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Namespace: seedName,
					Name:      cluster.Name,
				}}}
			}
		}
		return []reconcile.Request{}
	})
}

// EnqueueClusterScopedObjectWithSeedName enqueues a cluster-scoped object with the seedName
// as namespace. If it gets an object with a non-empty name, it will log an error and not enqueue
// anything.
func EnqueueClusterScopedObjectWithSeedName(seedName string) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(a ctrlruntimeclient.Object) []reconcile.Request {
		if a.GetNamespace() != "" {
			utilruntime.HandleError(fmt.Errorf("EnqueueClusterScopedObjectWithSeedName was used with namespace scoped object %s/%s of type %T", a.GetNamespace(), a.GetName(), a))
		}

		return []reconcile.Request{{NamespacedName: types.NamespacedName{
			Namespace: seedName,
			Name:      a.GetName(),
		}}}
	})
}

// EnqueueProjectForCluster returns an event handler that creates a reconcile
// request for the project of a cluster, based on the ProjectIDLabelKey label.
func EnqueueProjectForCluster() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(a ctrlruntimeclient.Object) []reconcile.Request {
		cluster, ok := a.(*kubermaticv1.Cluster)
		if !ok {
			return nil
		}

		projectID := cluster.Labels[kubermaticv1.ProjectIDLabelKey]
		if projectID == "" {
			return nil // should never happen, a webhook ensures a proper label
		}

		return []reconcile.Request{{NamespacedName: types.NamespacedName{
			Name: projectID,
		}}}
	})
}

const (
	CreateOperation = "create"
	UpdateOperation = "update"
	DeleteOperation = "delete"
)

// EnqueueObjectWithOperation enqueues a the namespaced name for any resource. It puts
// the current operation (create, update, delete) as the namespace and "namespace/name"
// of the object into the name of the reconcile request.
func EnqueueObjectWithOperation() handler.EventHandler {
	return handler.Funcs{
		CreateFunc: func(e event.CreateEvent, queue workqueue.RateLimitingInterface) {
			queue.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: CreateOperation,
					Name:      fmt.Sprintf("%s/%s", e.Object.GetNamespace(), e.Object.GetName()),
				},
			})
		},

		UpdateFunc: func(e event.UpdateEvent, queue workqueue.RateLimitingInterface) {
			queue.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: UpdateOperation,
					Name:      fmt.Sprintf("%s/%s", e.ObjectNew.GetNamespace(), e.ObjectNew.GetName()),
				},
			})
		},

		DeleteFunc: func(e event.DeleteEvent, queue workqueue.RateLimitingInterface) {
			queue.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: DeleteOperation,
					Name:      fmt.Sprintf("%s/%s", e.Object.GetNamespace(), e.Object.GetName()),
				},
			})
		},
	}
}

// EnqueueConst enqueues a constant. It is meant for controllers that don't have a parent object
// they could enc and instead reconcile everything at once.
// The queueKey will be defaulted if empty.
func EnqueueConst(queueKey string) handler.EventHandler {
	if queueKey == "" {
		queueKey = "const"
	}

	return handler.EnqueueRequestsFromMapFunc(func(o ctrlruntimeclient.Object) []reconcile.Request {
		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{
				Name:      queueKey,
				Namespace: "",
			}}}
	})
}

// ClusterAvailableForReconciling returns true if the given cluster can be reconciled. This is true if
// the cluster does not yet have the SeedResourcesUpToDate condition or if the concurrency limit of the
// controller is not yet reached. This ensures that not too many cluster updates are running at the same
// time, but also makes sure that un-UpToDate clusters will continue to be reconciled.
func ClusterAvailableForReconciling(ctx context.Context, client ctrlruntimeclient.Client, cluster *kubermaticv1.Cluster, concurrencyLimit int) (bool, error) {
	if !cluster.Status.HasConditionValue(kubermaticv1.ClusterConditionSeedResourcesUpToDate, corev1.ConditionTrue) {
		return true, nil
	}

	limitReached, err := ConcurrencyLimitReached(ctx, client, concurrencyLimit)
	return !limitReached, err
}

// ConcurrencyLimitReached checks all the clusters inside the seed cluster and checks for the
// SeedResourcesUpToDate condition. Returns true if the number of clusters without this condition
// is equal or larger than the given limit.
func ConcurrencyLimitReached(ctx context.Context, client ctrlruntimeclient.Client, limit int) (bool, error) {
	clusters := &kubermaticv1.ClusterList{}
	if err := client.List(ctx, clusters); err != nil {
		return true, fmt.Errorf("failed to list clusters: %w", err)
	}

	finishedUpdatingClustersCount := 0
	for _, cluster := range clusters.Items {
		if cluster.Status.HasConditionValue(kubermaticv1.ClusterConditionSeedResourcesUpToDate, corev1.ConditionTrue) {
			finishedUpdatingClustersCount++
		}
	}

	clustersUpdatingInProgressCount := len(clusters.Items) - finishedUpdatingClustersCount

	return clustersUpdatingInProgressCount >= limit, nil
}
