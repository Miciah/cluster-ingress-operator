package ingress

import (
	"context"
	"fmt"

	"github.com/openshift/cluster-ingress-operator/pkg/operator/controller"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const serviceCaConfiguration = `disablePermitInsecure: false
accesslog-format: envoy
`

// ensureServiceCAConfigMap ensures the ServiceCA configmap exists for a given
// ingresscontroller if it uses ServiceCA.  Returns a Boolean indicating whether
// the configmap exists, the configmap if it does exist, and an error value.
func (r *reconciler) ensureServiceCAConfigMap() (bool, *corev1.ConfigMap, error) {
	want, desired := desiredServiceCAConfigMap()

	have, current, err := r.currentServiceCAConfigMap()
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
			return true, current, fmt.Errorf("failed to delete configmap: %v", err)
		}
		log.Info("deleted configmap", "configmap", current)
	case want && !have:
		if err := r.client.Create(context.TODO(), desired); err != nil {
			return false, nil, fmt.Errorf("failed to create configmap: %v", err)
		}
		log.Info("created configmap", "configmap", desired)
	case want && have:
		// Nothing to do; the service CA updates the
		// configmap's data, and Contour and Envoy
		// automatically respond to changes.
	}

	return r.currentServiceCAConfigMap()
}

// desiredServiceCAConfigMap returns the desired service CA configmap.
// Returns a Boolean indicating whether a configmap is desired, as
// well as the configmap if one is desired.
func desiredServiceCAConfigMap() (bool, *corev1.ConfigMap) {
	name := controller.ServiceCAConfigMapName()
	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
			Annotations: map[string]string{
				"service.beta.openshift.io/inject-cabundle": "true",
			},
		},
	}

	return true, &cm
}

// currentServiceCAConfigMap returns the current service CA configmap.  Returns
// a Boolean indicating whether the configmap existed, the configmap if it did
// exist, and an error value.
func (r *reconciler) currentServiceCAConfigMap() (bool, *corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	if err := r.client.Get(context.TODO(), controller.ServiceCAConfigMapName(), cm); err != nil {
		if errors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, cm, nil
}
