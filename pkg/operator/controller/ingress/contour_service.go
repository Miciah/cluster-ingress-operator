package ingress

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-ingress-operator/pkg/manifests"
	"github.com/openshift/cluster-ingress-operator/pkg/operator/controller"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/intstr"
)

// ensureContourService ensures the Contour service exists for a
// given ingresscontroller if it uses Contour.  Returns a Boolean
// indicating whether the service exists, the service if it does
// exist, and an error value.
func (r *reconciler) ensureContourService(ic *operatorv1.IngressController, deploymentRef metav1.OwnerReference) (bool, *corev1.Service, error) {
	want, desired := desiredContourService(ic, deploymentRef)

	have, current, err := r.currentContourService(ic)
	if err != nil {
		return false, nil, err
	}

	switch {
	case !want && !have:
		return false, nil, nil
	case !want && have:
		if err := r.client.Delete(context.TODO(), current); err != nil {
			if errors.IsNotFound(err) {
				return false, nil, nil
			}
			return true, current, fmt.Errorf("failed to delete Contour service: %v", err)
		}
		log.Info("deleted Contour service", "service", current)
	case want && !have:
		if err := r.client.Create(context.TODO(), desired); err != nil {
			return false, nil, fmt.Errorf("failed to create Contour service: %v", err)
		}
		log.Info("created Contour service", "service", desired)
	case want && have:
		if updated, err := r.updateContourService(current, desired); err != nil {
			return true, nil, fmt.Errorf("failed to update service: %v", err)
		} else if updated {
			log.Info("updated service", "current service", current, "desired service", desired)
		}
	}

	return r.currentContourService(ic)
}

// desiredContourService returns the desired Contour service.  Returns a
// Boolean indicating whether a service is desired, as well as the service
// if one is desired.
func desiredContourService(ic *operatorv1.IngressController, deploymentRef metav1.OwnerReference) (bool, *corev1.Service) {
	if !UseContour(ic) {
		return false, nil
	}

	name := controller.ContourServiceName(ic)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: name.Namespace,
			Name:      name.Name,
			Annotations: map[string]string{
				"service.beta.openshift.io/serving-cert-secret-name": controller.ContourSecretName(ic).Name,
			},
			Labels: map[string]string{
				"app":                                  "contour",
				manifests.OwningIngressControllerLabel: ic.Name,
			},
			OwnerReferences: []metav1.OwnerReference{deploymentRef},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "xds",
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(8001),
					TargetPort: intstr.FromInt(8001),
				},
			},
			Selector:        controller.ContourPodSelector(ic).MatchLabels,
			SessionAffinity: corev1.ServiceAffinityNone,
			Type:            corev1.ServiceTypeClusterIP,
		},
	}

	return true, service
}

// currentContourService returns the current Contour service.  Returns a
// Boolean indicating whether the service existed, the service if it did
// exist, and an error value.
func (r *reconciler) currentContourService(ic *operatorv1.IngressController) (bool, *corev1.Service, error) {
	service := &corev1.Service{}
	if err := r.client.Get(context.TODO(), controller.ContourServiceName(ic), service); err != nil {
		if errors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, service, nil
}

// updateContourService updates the Contour service.  Returns a Boolean
// indicating whether the service was updated, and an error value.
func (r *reconciler) updateContourService(current, desired *corev1.Service) (bool, error) {
	changed, updated := contourServiceChanged(current, desired)
	if !changed {
		return false, nil
	}

	if err := r.client.Update(context.TODO(), updated); err != nil {
		return false, err
	}
	return true, nil
}

// contourServiceChanged checks if the current Contour service spec matches the
// expected spec and if not returns an updated one.
func contourServiceChanged(current, expected *corev1.Service) (bool, *corev1.Service) {
	serviceCmpOpts := []cmp.Option{
		// Ignore fields that the API, other controllers, or user may
		// have modified.
		cmpopts.IgnoreFields(corev1.ServicePort{}, "NodePort"),
		cmpopts.IgnoreFields(corev1.ServiceSpec{}, "ClusterIP", "ExternalIPs", "HealthCheckNodePort"),
		cmpopts.EquateEmpty(),
	}
	if cmp.Equal(current.Spec, expected.Spec, serviceCmpOpts...) {
		return false, nil
	}

	diff := cmp.Diff(current.Spec, expected.Spec, serviceCmpOpts...)
	log.Info("current service spec differs from expected", "namespace", current.Namespace, "name", current.Name, "diff", diff)

	updated := current.DeepCopy()
	updated.Spec = expected.Spec

	// Preserve fields that the API, other controllers, or user may have
	// modified.
	updated.Spec.ClusterIP = current.Spec.ClusterIP
	for i, updatedPort := range updated.Spec.Ports {
		for _, currentPort := range current.Spec.Ports {
			if currentPort.Name == updatedPort.Name {
				updated.Spec.Ports[i].Port = currentPort.Port
			}
		}
	}

	return true, updated
}
