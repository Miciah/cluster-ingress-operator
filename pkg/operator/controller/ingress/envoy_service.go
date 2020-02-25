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

// ensureEnvoyService ensures the Envoy service exists for a
// given ingresscontroller if it uses Envoy.  Returns a Boolean
// indicating whether the service exists, the service if it does
// exist, and an error value.
func (r *reconciler) ensureEnvoyService(ic *operatorv1.IngressController, deploymentRef metav1.OwnerReference) (bool, *corev1.Service, error) {
	want, desired := desiredEnvoyService(ic, deploymentRef)

	have, current, err := r.currentEnvoyService(ic)
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
			return true, current, fmt.Errorf("failed to delete Envoy service: %v", err)
		}
		log.Info("deleted Envoy service", "service", current)
	case want && !have:
		if err := r.client.Create(context.TODO(), desired); err != nil {
			return false, nil, fmt.Errorf("failed to create Envoy service: %v", err)
		}
		log.Info("created Envoy service", "service", desired)
	case want && have:
		if updated, err := r.updateEnvoyService(current, desired); err != nil {
			return true, nil, fmt.Errorf("failed to update service: %v", err)
		} else if updated {
			log.Info("updated service", "current service", current, "desired service", desired)
		}
	}

	return r.currentEnvoyService(ic)
}

// desiredEnvoyService returns the desired Envoy service.  Returns a
// Boolean indicating whether a service is desired, as well as the service
// if one is desired.
func desiredEnvoyService(ic *operatorv1.IngressController, deploymentRef metav1.OwnerReference) (bool, *corev1.Service) {
	if !UseContour(ic) {
		return false, nil
	}

	name := controller.EnvoyServiceName(ic)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: name.Namespace,
			Name:      name.Name,
			Annotations: map[string]string{
				"service.beta.openshift.io/serving-cert-secret-name": controller.EnvoySecretName(ic).Name,
			},
			Labels: map[string]string{
				"app":                                  "envoy",
				manifests.OwningIngressControllerLabel: ic.Name,
			},
			OwnerReferences: []metav1.OwnerReference{deploymentRef},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(80),
					TargetPort: intstr.FromInt(80),
				},
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(443),
					TargetPort: intstr.FromInt(443),
				},
			},
			Selector:        controller.EnvoyPodSelector(ic).MatchLabels,
			SessionAffinity: corev1.ServiceAffinityNone,
			Type:            corev1.ServiceTypeClusterIP,
		},
	}

	return true, service
}

// currentEnvoyService returns the current Envoy service.  Returns a
// Boolean indicating whether the service existed, the service if it did
// exist, and an error value.
func (r *reconciler) currentEnvoyService(ic *operatorv1.IngressController) (bool, *corev1.Service, error) {
	service := &corev1.Service{}
	if err := r.client.Get(context.TODO(), controller.EnvoyServiceName(ic), service); err != nil {
		if errors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, service, nil
}

// updateEnvoyService updates the Envoy service.  Returns a Boolean
// indicating whether the service was updated, and an error value.
func (r *reconciler) updateEnvoyService(current, desired *corev1.Service) (bool, error) {
	changed, updated := envoyServiceChanged(current, desired)
	if !changed {
		return false, nil
	}

	if err := r.client.Update(context.TODO(), updated); err != nil {
		return false, err
	}
	return true, nil
}

// envoyServiceChanged checks if the current Envoy service spec matches the
// expected spec and if not returns an updated one.
func envoyServiceChanged(current, expected *corev1.Service) (bool, *corev1.Service) {
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

	diff := cmp.Diff(current.Spec, expected.Spec, serviceCmpOpts...)
	log.Info("current service spec differs from expected", "namespace", current.Namespace, "name", current.Name, "diff", diff)

	return true, updated
}
