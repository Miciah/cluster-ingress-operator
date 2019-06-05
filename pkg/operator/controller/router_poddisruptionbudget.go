package controller

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	operatorv1 "github.com/openshift/api/operator/v1"

	policyv1beta1 "k8s.io/api/policy/v1beta1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ensureRouterPodDisruptionBudget ensures the pod disruption budget exists for
// a given ingresscontroller.  Returns a Boolean indicating whether the PDB
// exists, the PDB if it does exist, and an error value.
func (r *reconciler) ensureRouterPodDisruptionBudget(ic *operatorv1.IngressController, deploymentRef metav1.OwnerReference) (bool, *policyv1beta1.PodDisruptionBudget, error) {
	wantPDB, desired, err := desiredRouterPodDisruptionBudget(ic, deploymentRef)
	if err != nil {
		return false, nil, fmt.Errorf("failed to build pod disruption budget: %v", err)
	}

	havePDB, current, err := r.currentRouterPodDisruptionBudget(ic)
	if err != nil {
		return false, nil, err
	}

	switch {
	case !wantPDB && !havePDB:
		return false, nil, nil
	case !wantPDB && havePDB:
		if deleted, err := r.deleteRouterPodDisruptionBudget(current); err != nil {
			return true, current, fmt.Errorf("failed to delete pod disruption budget: %v", err)
		} else if deleted {
			log.Info("deleted pod disruption budget", "poddisruptionbudget", current)
		}
	case wantPDB && !havePDB:
		if created, err := r.createRouterPodDisruptionBudget(desired); err != nil {
			return false, nil, fmt.Errorf("failed to create pod disruption budget: %v", err)
		} else if created {
			log.Info("created pod disruption budget", "poddisruptionbudget", desired)
		}
	case wantPDB && havePDB:
		if updated, err := r.updateRouterPodDisruptionBudget(current, desired); err != nil {
			return true, nil, fmt.Errorf("failed to update pod disruption budget: %v", err)
		} else if updated {
			log.Info("updated pod disruption budget", "poddisruptionbudget", desired)
		}
	}

	return r.currentRouterPodDisruptionBudget(ic)
}

// desiredRouterPodDisruptionBudget returns the desired router pod disruption
// budget.  Returns a Boolean indicating whether a PDB is desired, as well as
// the PDB if one is desired.
func desiredRouterPodDisruptionBudget(ic *operatorv1.IngressController, deploymentRef metav1.OwnerReference) (bool, *policyv1beta1.PodDisruptionBudget, error) {
	name := RouterPodDisruptionBudgetName(ic)
	pointerTo := func(ios intstr.IntOrString) *intstr.IntOrString { return &ios }
	pdb := policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			// The disruption controller rounds MaxUnavailable up.
			// https://github.com/kubernetes/kubernetes/blob/65dc445aa2d581b4fa829258e46e4faf44e999b6/pkg/controller/disruption/disruption.go#L539
			MaxUnavailable: pointerTo(intstr.FromString("25%")),
			Selector:       IngressControllerDeploymentPodSelector(ic),
		},
	}
	pdb.SetOwnerReferences([]metav1.OwnerReference{deploymentRef})

	return true, &pdb, nil
}

// currentRouterPodDisruptionBudget returns the current router pod disruption
// budget.  Returns a Boolean indicating whether the PDB existed, the PDB if it
// did exist, and an error value.
func (r *reconciler) currentRouterPodDisruptionBudget(ic *operatorv1.IngressController) (bool, *policyv1beta1.PodDisruptionBudget, error) {
	pdb := &policyv1beta1.PodDisruptionBudget{}
	if err := r.client.Get(context.TODO(), RouterPodDisruptionBudgetName(ic), pdb); err != nil {
		if errors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, pdb, nil
}

// createRouterPodDisruptionBudget creates a pod disruption budget.  Returns a
// Boolean indicating whether the PDB was created, and an error value.
func (r *reconciler) createRouterPodDisruptionBudget(pdb *policyv1beta1.PodDisruptionBudget) (bool, error) {
	if err := r.client.Create(context.TODO(), pdb); err != nil {
		return false, err
	}
	log.Info("created pod disruption budget", "namespace", pdb.Namespace, "name", pdb.Name)
	return true, nil
}

// deleteRouterPodDisruptionBudget deletes a pod disruption budget.  Returns a
// Boolean indicating whether the PDB was deleted, and an error value.
func (r *reconciler) deleteRouterPodDisruptionBudget(pdb *policyv1beta1.PodDisruptionBudget) (bool, error) {
	if err := r.client.Delete(context.TODO(), pdb); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	log.Info("created pod disruption budget", "namespace", pdb.Namespace, "name", pdb.Name)
	return true, nil
}

// updateRouterPodDisruptionBudget updates a pod disruption budget.  Returns a
// Boolean indicating whether the PDB was updated, and an error value.
func (r *reconciler) updateRouterPodDisruptionBudget(current, desired *policyv1beta1.PodDisruptionBudget) (bool, error) {
	changed, updated := podDisruptionBudgetChanged(current, desired)
	if !changed {
		return false, nil
	}

	if err := r.client.Update(context.TODO(), updated); err != nil {
		return false, err
	}
	log.Info("updated pod disruption budget", "namespace", updated.Namespace, "name", updated.Name)
	return true, nil
}

// podDisruptionBudgetChanged checks if current pod disruption budget spec
// matches the expected spec and if not returns an updated one.
func podDisruptionBudgetChanged(current, expected *policyv1beta1.PodDisruptionBudget) (bool, *policyv1beta1.PodDisruptionBudget) {
	if cmp.Equal(current.Spec, expected.Spec, cmpopts.EquateEmpty()) {
		return false, nil
	}

	updated := current.DeepCopy()
	updated.Spec = expected.Spec
	return true, updated
}
