//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeviceSpec) DeepCopyInto(out *DeviceSpec) {
	*out = *in
	in.Selectors.DeepCopyInto(&out.Selectors)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeviceSpec.
func (in *DeviceSpec) DeepCopy() *DeviceSpec {
	if in == nil {
		return nil
	}
	out := new(DeviceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPUAddon) DeepCopyInto(out *GPUAddon) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPUAddon.
func (in *GPUAddon) DeepCopy() *GPUAddon {
	if in == nil {
		return nil
	}
	out := new(GPUAddon)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GPUAddon) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPUAddonList) DeepCopyInto(out *GPUAddonList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GPUAddon, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPUAddonList.
func (in *GPUAddonList) DeepCopy() *GPUAddonList {
	if in == nil {
		return nil
	}
	out := new(GPUAddonList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GPUAddonList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPUAddonSpec) DeepCopyInto(out *GPUAddonSpec) {
	*out = *in
	if in.RDMA != nil {
		in, out := &in.RDMA, &out.RDMA
		*out = new(RDMASpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPUAddonSpec.
func (in *GPUAddonSpec) DeepCopy() *GPUAddonSpec {
	if in == nil {
		return nil
	}
	out := new(GPUAddonSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GPUAddonStatus) DeepCopyInto(out *GPUAddonStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GPUAddonStatus.
func (in *GPUAddonStatus) DeepCopy() *GPUAddonStatus {
	if in == nil {
		return nil
	}
	out := new(GPUAddonStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPAMSpec) DeepCopyInto(out *IPAMSpec) {
	*out = *in
	if in.OmitRanges != nil {
		in, out := &in.OmitRanges, &out.OmitRanges
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPAMSpec.
func (in *IPAMSpec) DeepCopy() *IPAMSpec {
	if in == nil {
		return nil
	}
	out := new(IPAMSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MacvlanNetworkSpec) DeepCopyInto(out *MacvlanNetworkSpec) {
	*out = *in
	in.IPAM.DeepCopyInto(&out.IPAM)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MacvlanNetworkSpec.
func (in *MacvlanNetworkSpec) DeepCopy() *MacvlanNetworkSpec {
	if in == nil {
		return nil
	}
	out := new(MacvlanNetworkSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RDMASpec) DeepCopyInto(out *RDMASpec) {
	*out = *in
	if in.Devices != nil {
		in, out := &in.Devices, &out.Devices
		*out = make([]DeviceSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.MacvlanNetwork.DeepCopyInto(&out.MacvlanNetwork)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RDMASpec.
func (in *RDMASpec) DeepCopy() *RDMASpec {
	if in == nil {
		return nil
	}
	out := new(RDMASpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Selectors) DeepCopyInto(out *Selectors) {
	*out = *in
	if in.IfNames != nil {
		in, out := &in.IfNames, &out.IfNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Selectors.
func (in *Selectors) DeepCopy() *Selectors {
	if in == nil {
		return nil
	}
	out := new(Selectors)
	in.DeepCopyInto(out)
	return out
}
