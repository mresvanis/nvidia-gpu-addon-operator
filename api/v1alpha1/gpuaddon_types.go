/*
Copyright 2022.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Selectors contains common device selectors fields.
type Selectors struct {
	IfNames []string `json:"ifNames,omitempty"`
}

// DeviceSpec describes the user configuration for
// the the RDMA shared device plugin.
type DeviceSpec struct {
	// ResourceName describes the user selected device name.
	ResourceName string `json:"resourceName"`

	// Selectors describe the device selectors to be used.
	Selectors Selectors `json:"selectors"`
}

// IPAMSpec describes the configuration to be used for this network.
type IPAMSpec struct {
	Range string `json:"range"`

	OmitRanges []string `json:"exclude,omitempty"`
}

type MacvlanNetworkSpec struct {
	// Master describes the host interface to be used.
	Master string `json:"master,omitempty"`

	// IPAM configuration to be used for this network.
	IPAM IPAMSpec `json:"ipam,omitetmpy"`
}

// RDMASpec describes the configuration options of the
// NVIDIA Network Operator and the NVIDIA GPU Operator
// to enable GPUDirect RDMA.
type RDMASpec struct {
	Devices []DeviceSpec `json:"devices"`

	MacvlanNetwork MacvlanNetworkSpec `json:"macvlanNetwork"`
}

// GPUAddonSpec defines the desired state of GPUAddon.
type GPUAddonSpec struct {

	//+kubebuilder:default:=true
	// If enabled, addon will deploy the GPU console plugin.
	ConsolePluginEnabled bool `json:"console_plugin_enabled,omitempty"`

	// Optional RDMA configuration. If it's defined the operator will configure the
	// NVIDIA Network Operator and the NVIDIA GPU Operator resources, in order to
	// enable GPUDirect RDMA.
	RDMA *RDMASpec `json:"rdma,omitempty"`

	// Optional NVAIE pullsecret.
	NVAIEPullSecret string `json:"nvaie_pullsecret,omitempty"`
}

// GPUAddonStatus defines the observed state of GPUAddon.
type GPUAddonStatus struct {
	// The state of the addon operator.
	Phase GPUAddonPhase `json:"phase"`
	// Conditions represent the latest available observations of an object's state.
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:validation:Enum=Failed;Idle;Installing;Ready;Updating;Uninstalling
type GPUAddonPhase string

const (
	GPUAddonPhaseFailed       GPUAddonPhase = "Failed"
	GPUAddonPhaseIdle         GPUAddonPhase = "Idle"
	GPUAddonPhaseInstalling   GPUAddonPhase = "Installing"
	GPUAddonPhaseReady        GPUAddonPhase = "Ready"
	GPUAddonPhaseUpdating     GPUAddonPhase = "Updating"
	GPUAddonPhaseUninstalling GPUAddonPhase = "Uninstalling"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.addon_state`
//+kubebuilder:printcolumn:name="Console Plugin",type=boolean,JSONPath=`.spec.console_plugin_enabled`
//+kubebuilder:printcolumn:name="NVAIE State",type=string,JSONPath=`.status.nvaie_state`

// GPUAddon is the Schema for the gpuaddons API.
type GPUAddon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GPUAddonSpec   `json:"spec,omitempty"`
	Status GPUAddonStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GPUAddonList contains a list of GPUAddon.
type GPUAddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GPUAddon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GPUAddon{}, &GPUAddonList{})
}
