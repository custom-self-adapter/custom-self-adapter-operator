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

package controller

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"

	customselfadapternetv1 "github.com/custom-self-adapter/custom-self-adapter-operator/api/v1"
)

const (
	managedByLabel = "app.kubernetes.io/managed-by"
	OwnedByLabel   = "v1.custom-resources-adapter.net/owned-by"
)

type K8sReconciler interface {
	Reconcile(
		reqLogger logr.Logger,
		instance *customselfadapternetv1.CustomSelfAdapter,
		obj metav1.Object,
		shouldProvision bool,
		updateable bool,
		kind string,
	) (reconcile.Result, error)
	PodCleanup(reqLogger logr.Logger, instance *customselfadapternetv1.CustomSelfAdapter) error
}

// PrimaryPred is the predicate that filters events for the ScriptAdapter primary resource.
var PrimaryPred = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		return true
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		return true
	},
	CreateFunc: func(e event.CreateEvent) bool {
		return true
	},
	GenericFunc: func(e event.GenericEvent) bool {
		return false
	},
}

// SecondaryPred is the predicate that filters events for the ScriptAdapter's secondary
// resources (deployment/service/role/rolebinding).
var SecondaryPred = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		return false
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		return true
	},
	CreateFunc: func(e event.CreateEvent) bool {
		return false
	},
	GenericFunc: func(e event.GenericEvent) bool {
		return false
	},
}

// CustomSelfAdapterReconciler reconciles a CustomSelfAdapter object
type CustomSelfAdapterReconciler struct {
	client.Client
	Log                          logr.Logger
	Scheme                       *runtime.Scheme
	KubernetesResourceReconciler K8sReconciler
}

// +kubebuilder:rbac:groups=custom-self-adapter.net,resources=customselfadapters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=custom-self-adapter.net,resources=customselfadapters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=custom-self-adapter.net,resources=customselfadapters/finalizers,verbs=update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=*
// +kubebuilder:rbac:groups="",resources=replicationcontrollers;replicationcontrollers/scale,verbs=*
// +kubebuilder:rbac:groups=apps,resources=replicasets;replicasets/scale,verbs=*
// +kubebuilder:rbac:groups=apps,resources=statefulsets;statefulsets/scale,verbs=*
// +kubebuilder:rbac:groups=apps,resources=deployments;deployments/scale,verbs=*
// +kubebuilder:rbac:groups=metrics.k8s.io;custom.metrics.k8s.io;external.metrics.k8s.io,resources=*,verbs=*

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *CustomSelfAdapterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("Request", req.NamespacedName)

	instance := &customselfadapternetv1.CustomSelfAdapter{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if instance.DeletionTimestamp != nil {
		reqLogger.Info("Script K8S Adapter marked for deletion, ignoring reconciliation of dependencies ", "Kind", "script-k8s-adapter.xisberto.net/v1alpha1/ScriptAdapter", "Namespace", instance.GetNamespace(), "Name", instance.GetName())
		return reconcile.Result{}, nil
	}

	if instance.Spec.ProvisionRole == nil {
		defaultVal := true
		instance.Spec.ProvisionRole = &defaultVal
	}
	if instance.Spec.ProvisionRoleBinding == nil {
		defaultVal := true
		instance.Spec.ProvisionRoleBinding = &defaultVal
	}
	if instance.Spec.ProvisionServiceAccount == nil {
		defaultVal := true
		instance.Spec.ProvisionServiceAccount = &defaultVal
	}
	if instance.Spec.ProvisionPod == nil {
		defaultVal := true
		instance.Spec.ProvisionPod = &defaultVal
	}
	if instance.Spec.RoleRequiresMetricsServer == nil {
		defaultVal := false
		instance.Spec.RoleRequiresMetricsServer = &defaultVal
	}
	if instance.Spec.RoleRequiresArgoRollouts == nil {
		defaultVal := false
		instance.Spec.RoleRequiresArgoRollouts = &defaultVal
	}

	// Parse scaleTargetRef
	scaleTargetRef, err := json.Marshal(instance.Spec.ScaleTargetRef)
	if err != nil {
		panic(err)
	}

	labels := map[string]string{
		managedByLabel: "script-k8s-adapter-operator",
		OwnedByLabel:   instance.Name,
	}

	// Define and Reconcile a new Service Account object
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
	}
	result, err := r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, serviceAccount, *instance.Spec.ProvisionServiceAccount, true, "v1/ServiceAccount")
	if err != nil {
		return result, err
	}

	// Define a new Role Object
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "replicationcontrollers", "replicationcontrollers/scale"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "deployments/scale", "replicasets", "replicasets/scale", "statefulsets", "statefulsets/scale"},
				Verbs:     []string{"*"},
			},
		},
	}

	if *instance.Spec.RoleRequiresMetricsServer {
		role.Rules = append(role.Rules, rbacv1.PolicyRule{
			APIGroups: []string{"metrics.k8s.io", "custom.metrics.k8s.io", "external.metrics.k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		})
	}

	if *instance.Spec.RoleRequiresArgoRollouts {
		role.Rules = append(role.Rules, rbacv1.PolicyRule{
			APIGroups: []string{"argoproj.io"},
			Resources: []string{"rollouts", "rollouts/scale"},
			Verbs:     []string{"*"},
		})
	}
	result, err = r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, role, *instance.Spec.ProvisionRole, true, "v1/Role")
	if err != nil {
		return result, err
	}

	// Define and Reconcile a new Role Binding object
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      instance.Name,
				Namespace: instance.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     instance.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	result, err = r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, roleBinding, *instance.Spec.ProvisionRoleBinding, true, "v1/RoleBinding")
	if err != nil {
		return result, err
	}

	// Set up Pod labels, if labels are provided in the template Pod Spec the labels are merged
	// with the SKA managed-by label, otherwise only the managed-by label is added
	var podLabels map[string]string
	if instance.Spec.Template.ObjectMeta.Labels == nil {
		podLabels = map[string]string{}
	} else {
		podLabels = instance.Spec.Template.ObjectMeta.Labels
	}
	podLabels[managedByLabel] = "script-k8s-adapter-operator"
	podLabels[OwnedByLabel] = instance.Name

	// Set up ObjectMeta, if no name or namespaces are provided in the template PodSpec then
	// the SKA name and namespace are used
	objectMeta := instance.Spec.Template.ObjectMeta
	if objectMeta.Name == "" {
		objectMeta.Name = instance.Name
	}
	if objectMeta.Namespace == "" {
		objectMeta.Namespace = instance.Namespace
	}
	objectMeta.Labels = podLabels

	// Set up the PodSpec template
	podSpec := instance.Spec.Template.Spec
	// Inject environment variables to every Container specified by the PodSpec
	containers := []corev1.Container{}
	for _, container := range podSpec.Containers {
		// If no environment variables specified by the template PodSpec, set up basic env vars slice
		// Inject instance name and namespace to Env
		envVars := []corev1.EnvVar{
			corev1.EnvVar{
				Name:  "CSA_NAMESPACE",
				Value: instance.Namespace,
			},
			corev1.EnvVar{
				Name:  "CSA_NAME",
				Value: instance.Name,
			},
		}
		if container.Env != nil {
			envVars = append(envVars, container.Env...)
		}
		// Inject in configuration, such as namespace, target ref and configuration
		// options as environment variables
		envVars = append(envVars, csaEnvVars(instance, string(scaleTargetRef))...)
		container.Env = envVars
		containers = append(containers, container)
	}
	// Update PodSpec to use the modified containers, and to point to the provisioned service account
	podSpec.Containers = containers
	podSpec.ServiceAccountName = serviceAccount.Name

	// Define Pod object with ObjectMeta and modified PodSpec
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta(objectMeta),
		Spec:       corev1.PodSpec(podSpec),
	}
	result, err = r.KubernetesResourceReconciler.Reconcile(reqLogger, instance, pod, *instance.Spec.ProvisionPod, false, "v1/Pod")
	if err != nil {
		return result, err
	}

	// Clean up any orphaned pods (e.g. renaming pod, old pod should be deleted)
	err = r.KubernetesResourceReconciler.PodCleanup(reqLogger, instance)
	if err != nil {
		return result, err
	}

	return result, nil
}

func csaEnvVars(csa *customselfadapternetv1.CustomSelfAdapter, scaleTargetRef string) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  "scaleTargetRef",
			Value: scaleTargetRef,
		},
		{
			Name:  "namespace",
			Value: csa.Namespace,
		},
	}
	envVars = append(envVars, createEnvVarsFromConfig(csa.Spec.Config)...)
	return envVars
}

func createEnvVarsFromConfig(configs []customselfadapternetv1.CustomSelfAdapterConfig) []corev1.EnvVar {
	envVars := []corev1.EnvVar{}
	for _, config := range configs {
		envVars = append(envVars, corev1.EnvVar{
			Name:  config.Name,
			Value: config.Value,
		})
	}
	return envVars
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomSelfAdapterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&customselfadapternetv1.CustomSelfAdapter{}).
		Named("customselfadapter").
		WithEventFilter(PrimaryPred).
		Owns(&corev1.Pod{}, builder.WithPredicates(SecondaryPred)).
		Owns(&corev1.ServiceAccount{}, builder.WithPredicates(SecondaryPred)).
		Owns(&rbacv1.Role{}, builder.WithPredicates(SecondaryPred)).
		Owns(&rbacv1.RoleBinding{}, builder.WithPredicates(SecondaryPred)).
		Complete(r)
}
