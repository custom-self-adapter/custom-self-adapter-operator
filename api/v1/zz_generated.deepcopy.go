//go:build !ignore_autogenerated

/*
Copyright 2025 The Custom Self-Adapter Developers.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomSelfAdapter) DeepCopyInto(out *CustomSelfAdapter) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomSelfAdapter.
func (in *CustomSelfAdapter) DeepCopy() *CustomSelfAdapter {
	if in == nil {
		return nil
	}
	out := new(CustomSelfAdapter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomSelfAdapter) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomSelfAdapterConfig) DeepCopyInto(out *CustomSelfAdapterConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomSelfAdapterConfig.
func (in *CustomSelfAdapterConfig) DeepCopy() *CustomSelfAdapterConfig {
	if in == nil {
		return nil
	}
	out := new(CustomSelfAdapterConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomSelfAdapterList) DeepCopyInto(out *CustomSelfAdapterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CustomSelfAdapter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomSelfAdapterList.
func (in *CustomSelfAdapterList) DeepCopy() *CustomSelfAdapterList {
	if in == nil {
		return nil
	}
	out := new(CustomSelfAdapterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomSelfAdapterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomSelfAdapterSpec) DeepCopyInto(out *CustomSelfAdapterSpec) {
	*out = *in
	in.Template.DeepCopyInto(&out.Template)
	out.ScaleTargetRef = in.ScaleTargetRef
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = make([]CustomSelfAdapterConfig, len(*in))
		copy(*out, *in)
	}
	if in.ProvisionRole != nil {
		in, out := &in.ProvisionRole, &out.ProvisionRole
		*out = new(bool)
		**out = **in
	}
	if in.ProvisionRoleBinding != nil {
		in, out := &in.ProvisionRoleBinding, &out.ProvisionRoleBinding
		*out = new(bool)
		**out = **in
	}
	if in.ProvisionServiceAccount != nil {
		in, out := &in.ProvisionServiceAccount, &out.ProvisionServiceAccount
		*out = new(bool)
		**out = **in
	}
	if in.ProvisionPod != nil {
		in, out := &in.ProvisionPod, &out.ProvisionPod
		*out = new(bool)
		**out = **in
	}
	if in.RoleRequiresMetricsServer != nil {
		in, out := &in.RoleRequiresMetricsServer, &out.RoleRequiresMetricsServer
		*out = new(bool)
		**out = **in
	}
	if in.RoleRequiresArgoRollouts != nil {
		in, out := &in.RoleRequiresArgoRollouts, &out.RoleRequiresArgoRollouts
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomSelfAdapterSpec.
func (in *CustomSelfAdapterSpec) DeepCopy() *CustomSelfAdapterSpec {
	if in == nil {
		return nil
	}
	out := new(CustomSelfAdapterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomSelfAdapterStatus) DeepCopyInto(out *CustomSelfAdapterStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomSelfAdapterStatus.
func (in *CustomSelfAdapterStatus) DeepCopy() *CustomSelfAdapterStatus {
	if in == nil {
		return nil
	}
	out := new(CustomSelfAdapterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodMeta) DeepCopyInto(out *PodMeta) {
	*out = *in
	in.CreationTimestamp.DeepCopyInto(&out.CreationTimestamp)
	if in.DeletionTimestamp != nil {
		in, out := &in.DeletionTimestamp, &out.DeletionTimestamp
		*out = (*in).DeepCopy()
	}
	if in.DeletionGracePeriodSeconds != nil {
		in, out := &in.DeletionGracePeriodSeconds, &out.DeletionGracePeriodSeconds
		*out = new(int64)
		**out = **in
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.OwnerReferences != nil {
		in, out := &in.OwnerReferences, &out.OwnerReferences
		*out = make([]metav1.OwnerReference, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Finalizers != nil {
		in, out := &in.Finalizers, &out.Finalizers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ManagedFields != nil {
		in, out := &in.ManagedFields, &out.ManagedFields
		*out = make([]metav1.ManagedFieldsEntry, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodMeta.
func (in *PodMeta) DeepCopy() *PodMeta {
	if in == nil {
		return nil
	}
	out := new(PodMeta)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodSpec) DeepCopyInto(out *PodSpec) {
	*out = *in
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]corev1.Volume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.InitContainers != nil {
		in, out := &in.InitContainers, &out.InitContainers
		*out = make([]corev1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = make([]corev1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.EphemeralContainers != nil {
		in, out := &in.EphemeralContainers, &out.EphemeralContainers
		*out = make([]corev1.EphemeralContainer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.TerminationGracePeriodSeconds != nil {
		in, out := &in.TerminationGracePeriodSeconds, &out.TerminationGracePeriodSeconds
		*out = new(int64)
		**out = **in
	}
	if in.ActiveDeadlineSeconds != nil {
		in, out := &in.ActiveDeadlineSeconds, &out.ActiveDeadlineSeconds
		*out = new(int64)
		**out = **in
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.AutomountServiceAccountToken != nil {
		in, out := &in.AutomountServiceAccountToken, &out.AutomountServiceAccountToken
		*out = new(bool)
		**out = **in
	}
	if in.ShareProcessNamespace != nil {
		in, out := &in.ShareProcessNamespace, &out.ShareProcessNamespace
		*out = new(bool)
		**out = **in
	}
	if in.SecurityContext != nil {
		in, out := &in.SecurityContext, &out.SecurityContext
		*out = new(corev1.PodSecurityContext)
		(*in).DeepCopyInto(*out)
	}
	if in.ImagePullSecrets != nil {
		in, out := &in.ImagePullSecrets, &out.ImagePullSecrets
		*out = make([]corev1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]corev1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.HostAliases != nil {
		in, out := &in.HostAliases, &out.HostAliases
		*out = make([]corev1.HostAlias, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Priority != nil {
		in, out := &in.Priority, &out.Priority
		*out = new(int32)
		**out = **in
	}
	if in.DNSConfig != nil {
		in, out := &in.DNSConfig, &out.DNSConfig
		*out = new(corev1.PodDNSConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.ReadinessGates != nil {
		in, out := &in.ReadinessGates, &out.ReadinessGates
		*out = make([]corev1.PodReadinessGate, len(*in))
		copy(*out, *in)
	}
	if in.RuntimeClassName != nil {
		in, out := &in.RuntimeClassName, &out.RuntimeClassName
		*out = new(string)
		**out = **in
	}
	if in.EnableServiceLinks != nil {
		in, out := &in.EnableServiceLinks, &out.EnableServiceLinks
		*out = new(bool)
		**out = **in
	}
	if in.PreemptionPolicy != nil {
		in, out := &in.PreemptionPolicy, &out.PreemptionPolicy
		*out = new(corev1.PreemptionPolicy)
		**out = **in
	}
	if in.Overhead != nil {
		in, out := &in.Overhead, &out.Overhead
		*out = make(corev1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	if in.TopologySpreadConstraints != nil {
		in, out := &in.TopologySpreadConstraints, &out.TopologySpreadConstraints
		*out = make([]corev1.TopologySpreadConstraint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SetHostnameAsFQDN != nil {
		in, out := &in.SetHostnameAsFQDN, &out.SetHostnameAsFQDN
		*out = new(bool)
		**out = **in
	}
	if in.OS != nil {
		in, out := &in.OS, &out.OS
		*out = new(corev1.PodOS)
		**out = **in
	}
	if in.HostUsers != nil {
		in, out := &in.HostUsers, &out.HostUsers
		*out = new(bool)
		**out = **in
	}
	if in.SchedulingGates != nil {
		in, out := &in.SchedulingGates, &out.SchedulingGates
		*out = make([]corev1.PodSchedulingGate, len(*in))
		copy(*out, *in)
	}
	if in.ResourceClaims != nil {
		in, out := &in.ResourceClaims, &out.ResourceClaims
		*out = make([]corev1.PodResourceClaim, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodSpec.
func (in *PodSpec) DeepCopy() *PodSpec {
	if in == nil {
		return nil
	}
	out := new(PodSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodTemplateSpec) DeepCopyInto(out *PodTemplateSpec) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodTemplateSpec.
func (in *PodTemplateSpec) DeepCopy() *PodTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(PodTemplateSpec)
	in.DeepCopyInto(out)
	return out
}
