package ingress

import (
	"context"
	"fmt"
	"reflect"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-ingress-operator/pkg/operator/controller"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const contourConfiguration = `disablePermitInsecure: false
accesslog-format: envoy
`

// ensureContourConfigMap ensures the Contour configmap exists for a
// given ingresscontroller if it uses Contour.  Returns a Boolean
// indicating whether the configmap exists, the configmap if it does
// exist, and an error value.
func (r *reconciler) ensureContourConfigMap(ic *operatorv1.IngressController, deploymentRef metav1.OwnerReference) (bool, *corev1.ConfigMap, error) {
	want, desired := desiredContourConfigMap(ic, deploymentRef)

	have, current, err := r.currentContourConfigMap(ic)
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
		if updated, err := r.updateContourConfigMap(current, desired); err != nil {
			return true, nil, fmt.Errorf("failed to update configmap: %v", err)
		} else if updated {
			log.Info("updated configmap", "configmap", desired)
		}
	}

	return r.currentContourConfigMap(ic)
}

// desiredContourConfigMap returns the desired Contour configmap.  Returns a
// Boolean indicating whether a configmap is desired, as well as the configmap
// if one is desired.
func desiredContourConfigMap(ic *operatorv1.IngressController, deploymentRef metav1.OwnerReference) (bool, *corev1.ConfigMap) {
	if !UseContour(ic) {
		return false, nil
	}

	name := controller.ContourConfigMapName(ic)
	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Data: map[string]string{
			"contour.yaml": contourConfiguration,
		},
	}
	cm.SetOwnerReferences([]metav1.OwnerReference{deploymentRef})

	return true, &cm
}

// currentContourConfigMap returns the current Contour configmap.  Returns a
// Boolean indicating whether the configmap existed, the configmap if it did
// exist, and an error value.
func (r *reconciler) currentContourConfigMap(ic *operatorv1.IngressController) (bool, *corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	if err := r.client.Get(context.TODO(), controller.ContourConfigMapName(ic), cm); err != nil {
		if errors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, cm, nil
}

// updateContourConfigMap updates a configmap.  Returns a Boolean indicating
// whether the configmap was updated, and an error value.
func (r *reconciler) updateContourConfigMap(current, desired *corev1.ConfigMap) (bool, error) {
	if contourConfigmapsEqual(current, desired) {
		return false, nil
	}
	updated := current.DeepCopy()
	updated.Data = desired.Data
	if err := r.client.Update(context.TODO(), updated); err != nil {
		if errors.IsAlreadyExists(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// contourConfigmapsEqual compares two Contour configmaps.  Returns true if the
// configmaps should be considered equal for the purpose of determining whether
// an update is necessary, false otherwise
func contourConfigmapsEqual(a, b *corev1.ConfigMap) bool {
	if !reflect.DeepEqual(a.Data, b.Data) {
		return false
	}
	return true
}
