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

package machine

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	semverlib "github.com/Masterminds/semver/v3"

	clusterv1alpha1 "github.com/kubermatic/machine-controller/pkg/apis/cluster/v1alpha1"
	"github.com/kubermatic/machine-controller/pkg/cloudprovider/util"
	providerconfig "github.com/kubermatic/machine-controller/pkg/providerconfig/types"
	apiv1 "k8c.io/kubermatic/v2/pkg/api/v1"
	kubermaticv1 "k8c.io/kubermatic/v2/pkg/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/pkg/resources"
	"k8c.io/kubermatic/v2/pkg/resources/cloudconfig"
	"k8c.io/kubermatic/v2/pkg/validation"
	"k8c.io/kubermatic/v2/pkg/validation/nodeupdate"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Deployment returns a Machine Deployment object for the given Node Deployment spec.
func Deployment(c *kubermaticv1.Cluster, nd *apiv1.NodeDeployment, dc *kubermaticv1.Datacenter, keys []*kubermaticv1.UserSSHKey, data resources.CredentialsData) (*clusterv1alpha1.MachineDeployment, error) {
	md := &clusterv1alpha1.MachineDeployment{}

	if nd.Name != "" {
		md.Name = nd.Name
	} else {
		// GenerateName can be set only if Name is empty to avoid confusing error:
		// https://github.com/kubernetes/kubernetes/issues/32220
		md.GenerateName = fmt.Sprintf("%s-worker-", c.Name)
	}

	// Add Annotations to Machine Deployment
	md.Annotations = nd.Annotations

	md.Namespace = metav1.NamespaceSystem
	md.Finalizers = []string{metav1.FinalizerDeleteDependents}

	md.Spec.Selector.MatchLabels = map[string]string{
		"machine": fmt.Sprintf("md-%s-%s", c.Name, rand.String(10)),
	}
	md.Spec.Template.Labels = md.Spec.Selector.MatchLabels
	md.Spec.Template.Spec.Labels = nd.Spec.Template.Labels
	if md.Spec.Template.Spec.Labels == nil {
		md.Spec.Template.Spec.Labels = make(map[string]string)
	}
	md.Spec.Template.Spec.Labels["system/cluster"] = c.Name
	projectID, ok := c.Labels[kubermaticv1.ProjectIDLabelKey]
	if ok {
		md.Spec.Template.Spec.Labels["system/project"] = projectID
	}

	var taints []corev1.Taint
	for _, taint := range nd.Spec.Template.Taints {
		taints = append(taints, corev1.Taint{
			Value:  taint.Value,
			Key:    taint.Key,
			Effect: corev1.TaintEffect(taint.Effect),
		})
	}
	md.Spec.Template.Spec.Taints = taints

	// Create a copy to avoid changing the ND when changing the MD
	replicas := nd.Spec.Replicas
	md.Spec.Replicas = &replicas

	md.Spec.Template.Spec.Versions.Kubelet = nd.Spec.Template.Versions.Kubelet

	// Deprecated: This is not supported for 1.24 and higher and is blocked by
	// Validate for 1.24+. Can be removed once 1.23 support is dropped.
	if nd.Spec.DynamicConfig != nil && *nd.Spec.DynamicConfig {
		kubeletVersion, err := semverlib.NewVersion(nd.Spec.Template.Versions.Kubelet)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubelet version: %w", err)
		}

		md.Spec.Template.Spec.ConfigSource = &corev1.NodeConfigSource{
			ConfigMap: &corev1.ConfigMapNodeConfigSource{
				Namespace:        resources.KubeSystemNamespaceName,
				Name:             fmt.Sprintf("kubelet-config-%d.%d", kubeletVersion.Major(), kubeletVersion.Minor()),
				KubeletConfigKey: "kubelet",
			},
		}
	}

	if len(c.Spec.MachineNetworks) > 0 {
		// TODO(mrIncompetent): Rename this finalizer to not contain the word "kubermatic" (For whitelabeling purpose)
		md.Spec.Template.Annotations = map[string]string{
			"machine-controller.kubermatic.io/initializers": "ipam",
		}
	}

	if nd.Spec.Paused != nil {
		md.Spec.Paused = *nd.Spec.Paused
	}

	config, err := getProviderConfig(c, nd, dc, keys, data)
	if err != nil {
		return nil, err
	}

	err = getProviderOS(config, nd)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	md.Spec.Template.Spec.ProviderSpec.Value = &runtime.RawExtension{Raw: b}

	return md, nil
}

//gocyclo:ignore
func getProviderConfig(c *kubermaticv1.Cluster, nd *apiv1.NodeDeployment, dc *kubermaticv1.Datacenter, keys []*kubermaticv1.UserSSHKey, data resources.CredentialsData) (*providerconfig.Config, error) {
	config := providerconfig.Config{}
	config.SSHPublicKeys = make([]string, len(keys))
	for i, key := range keys {
		config.SSHPublicKeys[i] = key.Spec.PublicKey
	}

	var (
		cloudExt *runtime.RawExtension
		err      error
	)

	credentials, err := resources.GetCredentials(data)
	if err != nil {
		return nil, err
	}

	switch {
	case nd.Spec.Template.Cloud.AWS != nil && dc.Spec.AWS != nil:
		config.CloudProvider = providerconfig.CloudProviderAWS
		cloudExt, err = getAWSProviderSpec(c, nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.Azure != nil && dc.Spec.Azure != nil:
		config.CloudProvider = providerconfig.CloudProviderAzure
		cloudExt, err = getAzureProviderSpec(c, nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.VSphere != nil && dc.Spec.VSphere != nil:
		config.CloudProvider = providerconfig.CloudProviderVsphere

		// We use OverwriteCloudConfig for VSphere to ensure we always use the credentials
		// passed in via frontend for the cloud-provider functionality.
		overwriteCloudConfig, err := cloudconfig.CloudConfig(c, dc, credentials)
		if err != nil {
			return nil, err
		}
		config.OverwriteCloudConfig = &overwriteCloudConfig

		cloudExt, err = getVSphereProviderSpec(c, nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.Openstack != nil && dc.Spec.Openstack != nil:
		config.CloudProvider = providerconfig.CloudProviderOpenstack
		if err := validation.ValidateCreateNodeSpec(c, &nd.Spec.Template, dc); err != nil {
			return nil, err
		}

		cloudExt, err = getOpenstackProviderSpec(c, nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.Hetzner != nil && dc.Spec.Hetzner != nil:
		config.CloudProvider = providerconfig.CloudProviderHetzner
		cloudExt, err = getHetznerProviderSpec(c, nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.Digitalocean != nil && dc.Spec.Digitalocean != nil:
		config.CloudProvider = providerconfig.CloudProviderDigitalocean
		cloudExt, err = getDigitaloceanProviderSpec(c, nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.Packet != nil && dc.Spec.Packet != nil:
		config.CloudProvider = providerconfig.CloudProviderPacket
		cloudExt, err = getPacketProviderSpec(c, nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.GCP != nil && dc.Spec.GCP != nil:
		config.CloudProvider = providerconfig.CloudProviderGoogle
		cloudExt, err = getGCPProviderSpec(c, nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.Kubevirt != nil && dc.Spec.Kubevirt != nil:
		config.CloudProvider = providerconfig.CloudProviderKubeVirt
		cloudExt, err = getKubevirtProviderSpec(nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.Alibaba != nil && dc.Spec.Alibaba != nil:
		config.CloudProvider = providerconfig.CloudProviderAlibaba
		cloudExt, err = getAlibabaProviderSpec(c, nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.Anexia != nil && dc.Spec.Anexia != nil:
		config.CloudProvider = providerconfig.CloudProviderAnexia
		cloudExt, err = getAnexiaProviderSpec(nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.Nutanix != nil && dc.Spec.Nutanix != nil:
		config.CloudProvider = providerconfig.CloudProviderNutanix
		cloudExt, err = getNutanixProviderSpec(c, nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	case nd.Spec.Template.Cloud.VMwareCloudDirector != nil && dc.Spec.VMwareCloudDirector != nil:
		config.CloudProvider = providerconfig.CloudProviderVcloudDirector
		cloudExt, err = getVMwareCloudDirectorProviderSpec(c, nd.Spec.Template, dc)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown cloud provider or cloud provider mismatch between node and datacenter")
	}
	config.CloudProviderSpec = *cloudExt

	if config.Network == nil {
		config.Network = &providerconfig.NetworkConfig{}
	}

	switch {
	case c.IsIPv4Only():
		config.Network.IPFamily = util.IPv4
	case c.IsIPv6Only():
		config.Network.IPFamily = util.IPv6
	case c.IsDualStack():
		config.Network.IPFamily = util.DualStack
	default:
		config.Network.IPFamily = util.Unspecified
	}

	return &config, nil
}

func getProviderOS(config *providerconfig.Config, nd *apiv1.NodeDeployment) error {
	var (
		err   error
		osExt *runtime.RawExtension
	)

	// OS specifics
	switch {
	case nd.Spec.Template.OperatingSystem.Ubuntu != nil:
		config.OperatingSystem = providerconfig.OperatingSystemUbuntu
		osExt, err = getUbuntuOperatingSystemSpec(nd.Spec.Template)
		if err != nil {
			return err
		}
	case nd.Spec.Template.OperatingSystem.CentOS != nil:
		config.OperatingSystem = providerconfig.OperatingSystemCentOS
		osExt, err = getCentOSOperatingSystemSpec(nd.Spec.Template)
		if err != nil {
			return err
		}
	case nd.Spec.Template.OperatingSystem.SLES != nil:
		config.OperatingSystem = providerconfig.OperatingSystemSLES
		osExt, err = getSLESOperatingSystemSpec(nd.Spec.Template)
		if err != nil {
			return err
		}
	case nd.Spec.Template.OperatingSystem.RHEL != nil:
		config.OperatingSystem = providerconfig.OperatingSystemRHEL
		osExt, err = getRHELOperatingSystemSpec(nd.Spec.Template)
		if err != nil {
			return err
		}
	case nd.Spec.Template.OperatingSystem.Flatcar != nil:
		config.OperatingSystem = providerconfig.OperatingSystemFlatcar
		osExt, err = getFlatcarOperatingSystemSpec(nd.Spec.Template)
		if err != nil {
			return err
		}
	case nd.Spec.Template.OperatingSystem.RockyLinux != nil:
		config.OperatingSystem = providerconfig.OperatingSystemRockyLinux
		osExt, err = getRockyLinuxOperatingSystemSpec(nd.Spec.Template)
		if err != nil {
			return err
		}
	default:
		return errors.New("no machine os was provided")
	}
	config.OperatingSystemSpec = *osExt

	return nil
}

// Validate if the node deployment structure fulfills certain requirements. It returns node deployment with updated
// kubelet version if it wasn't specified.
func Validate(nd *apiv1.NodeDeployment, controlPlaneVersion *semverlib.Version) (*apiv1.NodeDeployment, error) {
	if nd.Spec.Template.Cloud.Openstack == nil &&
		nd.Spec.Template.Cloud.Digitalocean == nil &&
		nd.Spec.Template.Cloud.AWS == nil &&
		nd.Spec.Template.Cloud.Hetzner == nil &&
		nd.Spec.Template.Cloud.VSphere == nil &&
		nd.Spec.Template.Cloud.Azure == nil &&
		nd.Spec.Template.Cloud.Packet == nil &&
		nd.Spec.Template.Cloud.GCP == nil &&
		nd.Spec.Template.Cloud.Kubevirt == nil &&
		nd.Spec.Template.Cloud.Alibaba == nil &&
		nd.Spec.Template.Cloud.Anexia == nil &&
		nd.Spec.Template.Cloud.Nutanix == nil &&
		nd.Spec.Template.Cloud.VMwareCloudDirector == nil {
		return nil, fmt.Errorf("node deployment needs to have cloud provider data")
	}

	var (
		kubeletVersion *semverlib.Version
		err            error
	)

	if nd.Spec.Template.Versions.Kubelet != "" {
		kubeletVersion, err = semverlib.NewVersion(nd.Spec.Template.Versions.Kubelet)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubelet version: %w", err)
		}

		if err = nodeupdate.EnsureVersionCompatible(controlPlaneVersion, kubeletVersion); err != nil {
			return nil, err
		}
	} else {
		kubeletVersion = controlPlaneVersion
	}

	nd.Spec.Template.Versions.Kubelet = kubeletVersion.String()

	constraint124, err := semverlib.NewConstraint(">= 1.24")
	if err != nil {
		return nil, fmt.Errorf("failed to parse 1.24 constraint: %w", err)
	}

	if nd.Spec.DynamicConfig != nil && *nd.Spec.DynamicConfig && constraint124.Check(kubeletVersion) {
		return nil, errors.New("dynamic config cannot be configured for Kubernetes 1.24 or higher")
	}

	// The default
	allowedTaintEffects := sets.NewString(
		string(corev1.TaintEffectNoExecute),
		string(corev1.TaintEffectNoSchedule),
		string(corev1.TaintEffectPreferNoSchedule),
	)
	for _, taint := range nd.Spec.Template.Taints {
		if taint.Key == "" {
			return nil, errors.New("taint key must be set")
		}
		if taint.Value == "" {
			return nil, errors.New("taint value must be set")
		}
		if !allowedTaintEffects.Has(taint.Effect) {
			return nil, fmt.Errorf("taint effect '%s' not allowed. Allowed: %s", taint.Effect, strings.Join(allowedTaintEffects.List(), ", "))
		}
	}

	return nd, nil
}
