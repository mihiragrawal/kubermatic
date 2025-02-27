/*
Copyright 2021 The Kubermatic Kubernetes Platform contributors.

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

package azure

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-05-01/network"
	"github.com/Azure/go-autorest/autorest/to"

	kubermaticv1 "k8c.io/kubermatic/v2/pkg/apis/kubermatic/v1"
	kuberneteshelper "k8c.io/kubermatic/v2/pkg/kubernetes"
	"k8c.io/kubermatic/v2/pkg/provider"
)

const (
	defaultVNetCIDRIPv4 = "10.0.0.0/16"
	defaultVNetCIDRIPv6 = "fd00::/48"
)

func vnetName(cluster *kubermaticv1.Cluster) string {
	return resourceNamePrefix + cluster.Name
}

func reconcileVNet(ctx context.Context, clients *ClientSet, location string, cluster *kubermaticv1.Cluster, update provider.ClusterUpdater) (*kubermaticv1.Cluster, error) {
	if cluster.Spec.Cloud.Azure.VNetName == "" {
		cluster.Spec.Cloud.Azure.VNetName = vnetName(cluster)
	}

	var resourceGroup = cluster.Spec.Cloud.Azure.ResourceGroup
	if cluster.Spec.Cloud.Azure.VNetResourceGroup != "" {
		resourceGroup = cluster.Spec.Cloud.Azure.VNetResourceGroup
	}

	vnet, err := clients.Networks.Get(ctx, resourceGroup, cluster.Spec.Cloud.Azure.VNetName, "")
	if err != nil && !isNotFound(vnet.Response) {
		return cluster, err
	}

	// if we found a VNET, we can check for the ownership tag to determine
	// if the referenced VNET is owned by this cluster and should be reconciled
	if !isNotFound(vnet.Response) && !hasOwnershipTag(vnet.Tags, cluster) {
		return update(ctx, cluster.Name, func(updatedCluster *kubermaticv1.Cluster) {
			updatedCluster.Spec.Cloud.Azure.VNetName = cluster.Spec.Cloud.Azure.VNetName
		})
	}

	var cidrs []string
	if cluster.IsIPv4Only() || cluster.IsDualStack() {
		cidrs = append(cidrs, defaultVNetCIDRIPv4)
	}
	if cluster.IsIPv6Only() || cluster.IsDualStack() {
		cidrs = append(cidrs, defaultVNetCIDRIPv6)
	}
	target := targetVnet(cluster.Spec.Cloud, location, cluster.Name, cidrs)

	// check for attributes of the existing VNET and return early if all values are already
	// as expected. Since there are a lot of pointers in the network.VirtualNetwork struct, we need to
	// do a lot of "!= nil" checks so this does not panic.
	//
	// Attributes we check:
	// - Address space CIDR
	if !(vnet.VirtualNetworkPropertiesFormat != nil && vnet.VirtualNetworkPropertiesFormat.AddressSpace != nil &&
		reflect.DeepEqual(vnet.VirtualNetworkPropertiesFormat.AddressSpace.AddressPrefixes, target.VirtualNetworkPropertiesFormat.AddressSpace.AddressPrefixes)) {
		if err := ensureVNet(ctx, clients, cluster.Spec.Cloud, target); err != nil {
			return nil, err
		}
	}

	return update(ctx, cluster.Name, func(updatedCluster *kubermaticv1.Cluster) {
		updatedCluster.Spec.Cloud.Azure.VNetName = cluster.Spec.Cloud.Azure.VNetName
		kuberneteshelper.AddFinalizer(updatedCluster, FinalizerVNet)
	})
}

func targetVnet(cloud kubermaticv1.CloudSpec, location string, clusterName string, cidrs []string) *network.VirtualNetwork {
	return &network.VirtualNetwork{
		Name:     to.StringPtr(cloud.Azure.VNetName),
		Location: to.StringPtr(location),
		Tags: map[string]*string{
			clusterTagKey: to.StringPtr(clusterName),
		},
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{AddressPrefixes: &cidrs},
		},
	}
}

// ensureVNet will create or update an Azure virtual network in the specified resource group. The call is idempotent.
func ensureVNet(ctx context.Context, clients *ClientSet, cloud kubermaticv1.CloudSpec, vnet *network.VirtualNetwork) error {
	if vnet == nil {
		return fmt.Errorf("invalid vnet reference passed")
	}

	var resourceGroup = cloud.Azure.ResourceGroup
	if cloud.Azure.VNetResourceGroup != "" {
		resourceGroup = cloud.Azure.VNetResourceGroup
	}
	future, err := clients.Networks.CreateOrUpdate(ctx, resourceGroup, cloud.Azure.VNetName, *vnet)
	if err != nil {
		return fmt.Errorf("failed to create or update virtual network %q: %w", cloud.Azure.VNetName, err)
	}

	if err = future.WaitForCompletionRef(ctx, *clients.Autorest); err != nil {
		return fmt.Errorf("failed to create or update virtual network %q: %w", cloud.Azure.VNetName, err)
	}

	return nil
}

func deleteVNet(ctx context.Context, clients *ClientSet, cloud kubermaticv1.CloudSpec) error {
	var resourceGroup = cloud.Azure.ResourceGroup
	if cloud.Azure.VNetResourceGroup != "" {
		resourceGroup = cloud.Azure.VNetResourceGroup
	}

	// We first do Get to check existence of the VNet to see if its already gone or not.
	// We could also directly call delete but the error response would need to be unpacked twice to get the correct error message.
	res, err := clients.Networks.Get(ctx, resourceGroup, cloud.Azure.VNetName, "")
	if err != nil {
		if isNotFound(res.Response) {
			return nil
		}
		return err
	}

	deleteVNetFuture, err := clients.Networks.Delete(ctx, resourceGroup, cloud.Azure.VNetName)
	if err != nil {
		return err
	}

	if err = deleteVNetFuture.WaitForCompletionRef(ctx, *clients.Autorest); err != nil {
		return err
	}

	return nil
}
