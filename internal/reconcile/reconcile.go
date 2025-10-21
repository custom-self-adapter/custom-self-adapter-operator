/*
Copyright 2022 The Custom Pod Autoscaler Authors.
Copyright 2025 The Script K8S Adapter Authors.

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

package reconcile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"maps"
	"strings"

	customselfadapternetv1 "github.com/custom-self-adapter/custom-self-adapter-operator/api/v1"
	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type controllerReferencer func(owner, object metav1.Object, scheme *runtime.Scheme, opts ...controllerutil.OwnerReferenceOption) error

// KubernetesResourceReconciler handles reconciling Kubernetes resources, such as pods, service accounts etc.
type KubernetesResourceReconciler struct {
	Scheme               *runtime.Scheme
	Client               client.Client
	ControllerReferencer controllerReferencer
}

// Reconcile manages k8s objects, making sure that the supplied object exists, and if it
// doesn't it creates one
func (k *KubernetesResourceReconciler) Reconcile(
	reqLogger logr.Logger,
	instance *customselfadapternetv1.CustomSelfAdapter,
	obj metav1.Object,
	shouldProvision bool,
	updatable bool,
	kind string,
) (reconcile.Result, error) {
	runtimeObj := obj.(client.Object)
	// Set CustomSelfAdapter instance as the owner and controller
	err := k.ControllerReferencer(instance, obj, k.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if k8s object already exists
	existingObject := runtimeObj
	err = k.Client.Get(context.Background(), types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, existingObject)
	if err != nil {
		if !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		// Object does not exist
		if !shouldProvision {
			reqLogger.Info("Object not found, no provisioning of resource ", "Kind", kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
			// Should not provision a new object, wait for existing
			return reconcile.Result{}, nil
		}
		// Should provision, create a new object
		reqLogger.Info("Creating a new k8s object ", "Kind", kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		err = k.Client.Create(context.Background(), runtimeObj)
		if err != nil {
			return reconcile.Result{}, err
		}
		// K8s object created successfully - don't requeue
		return reconcile.Result{}, nil
	}

	// TODO: Remove all Pod related code from Reconcile
	if existingObject.GetObjectKind().GroupVersionKind().Group == "" &&
		existingObject.GetObjectKind().GroupVersionKind().Version == "v1" &&
		existingObject.GetObjectKind().GroupVersionKind().Kind == "Pod" {
		pod := existingObject.(*corev1.Pod)
		if !pod.ObjectMeta.DeletionTimestamp.IsZero() {
			reqLogger.Info("Pod currently being deleted ", "Kind", kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
			return reconcile.Result{}, nil
		}
	}

	// Object already exists, update
	if shouldProvision {
		// Only update if object should be provisioned
		if updatable {
			reqLogger.Info("Updating k8s object ", "Kind", kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
			if existingObject.GetObjectKind().GroupVersionKind().Group == "" &&
				existingObject.GetObjectKind().GroupVersionKind().Version == "v1" &&
				existingObject.GetObjectKind().GroupVersionKind().Kind == "ServiceAccount" {
				reqLogger.Info("Service Account update, retaining secrets ", "Kind", kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
				serviceAccount := existingObject.(*corev1.ServiceAccount)
				updatedServiceAccount := runtimeObj.(*corev1.ServiceAccount)
				updatedServiceAccount.Secrets = serviceAccount.Secrets
			}
			// If object can be updated
			err = k.Client.Update(context.Background(), runtimeObj)
			if err != nil {
				return reconcile.Result{}, err
			}
			// Successful update, don't requeue
			return reconcile.Result{}, nil
		}
		reqLogger.Info("Deleting k8s object ", "Kind", kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())

		// If object can't be updated, delete and make new
		err = k.Client.Delete(context.Background(), existingObject)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	// Object should not be provisioned, instead update owner reference of
	// existing object
	obj = existingObject.(metav1.Object)
	// Check if SKA is set as K8s object owner
	ownerReferences := obj.GetOwnerReferences()
	csaOner := false
	for _, owner := range ownerReferences {
		if owner.Kind == instance.Kind && owner.APIVersion == instance.APIVersion && owner.Name == instance.Name {
			csaOner = true
			break
		}
	}

	if !csaOner {
		reqLogger.Info("CSA not set as owner, updating owner reference", "Kind", kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		ownerReferences = append(ownerReferences, metav1.OwnerReference{
			APIVersion: instance.APIVersion,
			Kind:       instance.Kind,
			Name:       instance.Name,
			UID:        instance.UID,
		})
		obj.SetOwnerReferences(ownerReferences)
		err = k.Client.Update(context.Background(), existingObject)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Skip reconcile: k8s object already exists with expected owner", "Kind", kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
	return reconcile.Result{}, nil
}

// ReconcilePod ensures a single, drift-aware Pod for the given owner:
//   - Create when missing
//   - Recreate when the desired template hash changes (immutable spec change)
//   - Patch only metadata (managed labels/annotations) when they diverge
//   - Call 'statusSync' to map Pod annotations -> CR.Status and clear transient annotations
func (r *KubernetesResourceReconciler) ReconcilePod(
	ctx context.Context,
	log logr.Logger,
	owner client.Object,
	desired *corev1.Pod,
	managedAnnPrefix string,
	templateHashKey string,
	ensureLabels map[string]string,
) (reconcile.Result, error) {

	if desired.Labels == nil {
		desired.Labels = map[string]string{}
	}
	// Ensure required labels (e.g., managed-by, owned-by)
	for k, v := range ensureLabels {
		if desired.Labels[k] != v {
			desired.Labels[k] = v
		}
	}

	if desired.Annotations == nil {
		desired.Annotations = map[string]string{}
	}

	// Hash desired template (managed metadata + spec) to detect drift
	tplHash, err := hashPodTemplate(desired.ObjectMeta, desired.Spec, managedAnnPrefix)
	if err != nil {
		return reconcile.Result{}, err
	}
	desired.Annotations[templateHashKey] = tplHash

	// OwnerRef for GC and event wiring
	if err := controllerutil.SetControllerReference(owner, desired, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Fetch current
	current := &corev1.Pod{}
	key := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
	if err := r.Client.Get(ctx, key, current); errors.IsNotFound(err) {
		log.Info("Creating Pod", "name", desired.Name, "ns", desired.Namespace)
		if err := r.Client.Create(ctx, desired); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// If terminating, skip now
	if current.DeletionTimestamp != nil && !current.DeletionTimestamp.IsZero() {
		log.Info("Pod is terminating; skip reconcile", "pod", current.Name)
		return reconcile.Result{}, nil
	}

	// 1) Map Pod annotations -> CSA.Status (via callback) and clear transient annotations
	if toClear, changed := statusSync(current, owner, managedAnnPrefix); changed {
		if err := r.Client.Status().Update(ctx, owner); err != nil {
			log.Error(err, "failed to update CR status")
			return reconcile.Result{}, err
		}
		if len(toClear) > 0 {
			orig := current.DeepCopy()
			if current.Annotations == nil {
				current.Annotations = map[string]string{}
			}
			for _, k := range toClear {
				delete(current.Annotations, k)
			}
			if err := r.Client.Patch(ctx, current, client.MergeFrom(orig)); err != nil {
				return reconcile.Result{}, err
			}
			log.Info("Cleared transient Pod annotations", "keys", toClear)
		}
	}

	// 2) Template drift check (immutable changes -> recreate)
	curHash := ""
	if current.Annotations != nil {
		curHash = current.Annotations[templateHashKey]
	}
	if curHash != tplHash {
		log.Info("Template hash changed; recreating Pod", "old", curHash, "new", tplHash)
		if err := r.Client.Delete(ctx, current); err != nil {
			return reconcile.Result{}, err
		}
		// Next reconcile (from Delete event) will create the new Pod
		return reconcile.Result{}, nil
	}

	// 3) Keep metadata in sync (ONLY what we manage)
	needPatch := false

	if current.Labels == nil {
		current.Labels = map[string]string{}
	}
	for k, v := range ensureLabels {
		if current.Labels[k] != v {
			current.Labels[k] = v
			needPatch = true
		}
	}

	if current.Annotations == nil {
		current.Annotations = map[string]string{}
	}
	desiredManaged := filterAnnotationsByPrefix(desired.Annotations, managedAnnPrefix)
	if anns, changed := syncManagedAnnotations(current.Annotations, desiredManaged, managedAnnPrefix); changed {
		current.Annotations = anns
		needPatch = true
	}

	if needPatch {
		orig := current.DeepCopy()
		if err := r.Client.Patch(ctx, current, client.MergeFrom(orig)); err != nil {
			return reconcile.Result{}, err
		}
		log.Info("Patched Pod metadata")
	}

	return reconcile.Result{}, nil
}

// --- helpers (package-private) ---

// hashPodTemplate computes a stable hash from the desired Pod template (managed metadata + spec).
// Only include fields representing intent; avoid volatile fields (UID, RV, timestamps).
func hashPodTemplate(meta metav1.ObjectMeta, spec corev1.PodSpec, managedAnnPrefix string) (string, error) {
	type podT struct {
		Meta struct {
			Name        string            `json:"name"`
			Labels      map[string]string `json:"labels,omitempty"`
			Annotations map[string]string `json:"annotations,omitempty"`
		} `json:"meta"`
		Spec corev1.PodSpec `json:"spec"`
	}

	pt := podT{}
	pt.Meta.Name = meta.Name
	pt.Meta.Labels = make(map[string]string, len(meta.Labels))
	maps.Copy(pt.Meta.Labels, meta.Labels)
	pt.Meta.Annotations = filterAnnotationsByPrefix(meta.Annotations, managedAnnPrefix)
	pt.Spec = spec

	b, err := json.Marshal(pt)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// statusSync lets the controller map Pod annotations into CSA.status fields.
// It mutates 'owner' (the CSA instance) in-memory and indicate whether status changed,
// returning which Pod annotations should be cleared afterwards.
func statusSync(pod *corev1.Pod, owner client.Object, managedAnnPrefix string) ([]string, bool) {
	csa := owner.(*customselfadapternetv1.CustomSelfAdapter)
	toClear := []string{}
	changed := false

	annInitialData := strings.Join([]string{managedAnnPrefix, "initialData"}, "/")
	if v := pod.GetAnnotations()[annInitialData]; v != "" && csa.Status.InitialData != v {
		csa.Status.InitialData = v
		toClear = append(toClear, annInitialData)
		changed = true
	}

	// TODO: add other annotations -> status mappings

	return toClear, changed
}

// syncManagedAnnotations applies desired managed annotations and removes extra managed keys.
// It never touches keys outside your managed prefix.
func syncManagedAnnotations(current map[string]string, desiredManaged map[string]string, prefix string) (map[string]string, bool) {
	changed := false
	out := make(map[string]string, len(current))
	maps.Copy(out, current)

	for k, v := range desiredManaged {
		if out[k] != v {
			out[k] = v
			changed = true
		}
	}
	for k := range out {
		if strings.HasPrefix(k, prefix) {
			if _, keep := desiredManaged[k]; !keep {
				delete(out, k)
				changed = true
			}
		}
	}
	return out, changed
}

func filterAnnotationsByPrefix(in map[string]string, prefix string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		if strings.HasPrefix(k, prefix) {
			out[k] = v
		}
	}
	return out
}

// PodCleanup will look for any Pods that have the v1.custompodautoscaler.com/owned-by label set to the name of the SKA
// and delete any 'orphaned' Pods, these are Pods that are owned by the SKA but are no longer defined in the SKA
// PodTemplateSpec (for example if the PodTemplateSpec has renamed the Pod, it should delete the old Pod as it
// provisions a new Pod so there aren't two Pods for the SKA)
func (k *KubernetesResourceReconciler) PodCleanup(reqLogger logr.Logger, instance *customselfadapternetv1.CustomSelfAdapter, ownedByLabel string) error {
	pods := &corev1.PodList{}
	err := k.Client.List(context.Background(), pods,
		client.MatchingLabels{ownedByLabel: instance.Name},
		client.InNamespace(instance.Namespace))

	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		managed := false
		for _, ownerRef := range pod.OwnerReferences {
			if ownerRef.APIVersion != instance.APIVersion || ownerRef.Kind != instance.Kind || ownerRef.Name != instance.Name {
				continue
			}

			managed = true
		}

		if !managed {
			continue
		}

		if instance.Spec.Template.ObjectMeta.Name == "" {
			// Using instance name, delete any pod that isn't using the instance name
			if pod.Name == instance.Name {
				continue
			}

			err = k.deleteOrphan(reqLogger, pod)
			if err != nil {
				return err
			}
			continue
		}

		// Using name defined in template, delete any pod that doesn't match that name
		if pod.Name != instance.Spec.Template.ObjectMeta.Name {
			err = k.deleteOrphan(reqLogger, pod)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (k *KubernetesResourceReconciler) deleteOrphan(reqLogger logr.Logger, pod corev1.Pod) error {
	reqLogger.Info("Found orphaned Pod (owned by CSA but not currently defined), deleting", "Kind", pod.GetObjectKind().GroupVersionKind(), "Namespace", pod.GetNamespace(), "Name", pod.GetName())
	return k.Client.Delete(context.Background(), &pod)
}
