package ingress

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"

	"github.com/google/go-cmp/cmp"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-ingress-operator/pkg/manifests"
	"github.com/openshift/cluster-ingress-operator/pkg/operator/controller"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"

	configv1 "github.com/openshift/api/config/v1"
)

// ensureEnvoyDaemonSet ensures the Envoy daemonset exists for a given
// ingresscontroller.
func (r *reconciler) ensureEnvoyDaemonSet(ic *operatorv1.IngressController, infraConfig *configv1.Infrastructure, ingressConfig *configv1.Ingress, apiConfig *configv1.APIServer, networkConfig *configv1.Network) (bool, *appsv1.DaemonSet, error) {
	want, desired, err := desiredEnvoyDaemonSet(ic, r.Config.IngressControllerImage, infraConfig, ingressConfig, apiConfig, networkConfig)
	if err != nil {
		return false, nil, err
	}

	have, current, err := r.currentEnvoyDaemonSet(ic)
	if err != nil {
		return false, nil, err
	}

	switch {
	case !want && !have:
		return false, nil, nil
	case !want && have:
		if err := r.deleteEnvoyDaemonSet(current); err != nil {
			return false, nil, err
		}
	case want && !have:
		if err := r.createEnvoyDaemonSet(desired); err != nil {
			return false, nil, err
		}
	case want && have:
		if err := r.updateEnvoyDaemonSet(current, desired); err != nil {
			return false, nil, err
		}
	}

	return r.currentEnvoyDaemonSet(ic)
}

// desiredEnvoyDaemonSet returns the desired Envoy daemonset.
func desiredEnvoyDaemonSet(ic *operatorv1.IngressController, ingressControllerImage string, infraConfig *configv1.Infrastructure, ingressConfig *configv1.Ingress, apiConfig *configv1.APIServer, networkConfig *configv1.Network) (bool, *appsv1.DaemonSet, error) {
	if !UseContour(ic) {
		return false, nil, nil
	}

	if ic.DeletionTimestamp != nil {
		return false, nil, nil
	}

	nodeSelector := map[string]string{
		"kubernetes.io/os":               "linux",
		"node-role.kubernetes.io/worker": "",
	}
	tolerations := []corev1.Toleration{}
	if ic.Spec.NodePlacement != nil {
		if ic.Spec.NodePlacement.NodeSelector != nil {
			var err error

			nodeSelector, err = metav1.LabelSelectorAsMap(ic.Spec.NodePlacement.NodeSelector)
			if err != nil {
				return false, nil, fmt.Errorf("ingresscontroller %q has invalid spec.nodePlacement.nodeSelector: %v", ic.Name, err)
			}
		}

		if ic.Spec.NodePlacement.Tolerations != nil {
			tolerations = ic.Spec.NodePlacement.Tolerations
		}
	}

	hostNetwork := false
	if ic.Status.EndpointPublishingStrategy.Type == operatorv1.HostNetworkStrategyType {
		hostNetwork = true
	}

	fifteenPercent := intstr.FromString("15%")
	fourTwenty := int32(420) // = 0644 octal.
	name := controller.EnvoyDaemonSetName(ic)
	daemonset := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
			Labels: map[string]string{
				manifests.OwningIngressControllerLabel: ic.Name,
			},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: controller.EnvoyPodSelector(ic),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "8002",
						"prometheus.io/path":   "/stats/prometheus",
					},
					Labels: controller.EnvoyPodSelector(ic).MatchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Args: []string{
								"-c",
								"/config/envoy.json",
								"--service-cluster $(CONTOUR_NAMESPACE)",
								"--service-node $(ENVOY_POD_NAME)",
								"--log-level info",
							},
							Command: []string{"envoy"},
							Env: []corev1.EnvVar{
								{
									Name: "CONTOUR_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
								{
									Name: "ENVOY_POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
							},
							Image:           "docker.io/envoyproxy/envoy:v1.12.2",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"bash",
											"-c",
											"--",
											"echo",
											"-ne",
											"POST /healthcheck/fail HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n",
											">/dev/tcp/localhost/9001",
										},
									},
								},
							},
							Name: "envoy",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(80),
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
								},
								{
									ContainerPort: int32(443),
									Name:          "https",
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Host: "localhost",
										Path: "/ready",
										Port: intstr.FromInt(9001),
									},
								},
								InitialDelaySeconds: int32(3),
								PeriodSeconds:       int32(3),
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cacert",
									MountPath: "/ca",
									ReadOnly:  true,
								},
								{
									Name:      "envoy-config",
									MountPath: "/config",
									ReadOnly:  true,
								},
								{
									Name:      "envoycert",
									MountPath: "/certs",
									ReadOnly:  true,
								},
							},
						},
					},
					DNSPolicy:   corev1.DNSClusterFirst,
					HostNetwork: hostNetwork,
					NodeSelector:       nodeSelector,
					ServiceAccountName: "router",
					Tolerations:        tolerations,
					Volumes: []corev1.Volume{
						{
							Name: "cacert",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: controller.ServiceCAConfigMapName().Name,
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "service-ca.crt",
											Path: "cacert.pem",
										},
									},
								},
							},
						},
						{
							Name: "envoy-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									DefaultMode: &fourTwenty,
									Items: []corev1.KeyToPath{
										{
											Key:  "envoy.json",
											Path: "envoy.json",
										},
									},
									LocalObjectReference: corev1.LocalObjectReference{
										Name: controller.EnvoyConfigMapName(ic).Name,
									},
								},
							},
						},
						{
							Name: "envoycert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: controller.EnvoySecretName(ic).Name,
								},
							},
						},
					},
				},
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: &fifteenPercent,
				},
			},
		},
	}

	return true, daemonset, nil
}

// daemonsetHash returns a stringified hash value for the Envoy daemonset
// fields that, if changed, should trigger an update.
func daemonsetHash(daemonset *appsv1.DaemonSet) string {
	hasher := fnv.New32a()
	deepHashObject(hasher, hashableDaemonSet(daemonset))
	return rand.SafeEncodeString(fmt.Sprint(hasher.Sum32()))
}

// hashableDaemonSet returns a copy of the given daemonset with exactly the
// fields from daemonset that should be used for computing its hash copied
// over.  In particular, these are the fields that desiredEnvoyDaemonSet sets.
// Fields with slice values will be sorted.  Fields that should be ignored, or
// that have explicit values that are equal to their respective default values,
// will be zeroed.
func hashableDaemonSet(daemonset *appsv1.DaemonSet) *appsv1.DaemonSet {
	var hashableDaemonSet appsv1.DaemonSet

	hashableDaemonSet.Labels = daemonset.Labels
	hashableDaemonSet.Name = daemonset.Name
	hashableDaemonSet.Namespace = daemonset.Namespace
	hashableDaemonSet.Spec.Template.Labels = daemonset.Spec.Template.Labels
	containers := make([]corev1.Container, len(daemonset.Spec.Template.Spec.Containers))
	for i, container := range daemonset.Spec.Template.Spec.Containers {
		env := container.Env
		sort.Slice(env, func(i, j int) bool {
			return env[i].Name < env[j].Name
		})
		ports := container.Ports
		sort.Slice(ports, func(i, j int) bool {
			return ports[i].Name < ports[j].Name
		})
		containers[i] = corev1.Container{
			Args:            container.Args,
			Command:         container.Command,
			Env:             env,
			Image:           container.Image,
			ImagePullPolicy: container.ImagePullPolicy,
			Name:            container.Name,
			Ports:           ports,
			VolumeMounts:    container.VolumeMounts,
		}
	}
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})
	hashableDaemonSet.Spec.Template.Spec.Containers = containers
	hashableDaemonSet.Spec.Template.Spec.HostNetwork = daemonset.Spec.Template.Spec.HostNetwork
	hashableDaemonSet.Spec.Template.Spec.NodeSelector = daemonset.Spec.Template.Spec.NodeSelector
	tolerations := make([]corev1.Toleration, len(daemonset.Spec.Template.Spec.Tolerations))
	for i, toleration := range daemonset.Spec.Template.Spec.Tolerations {
		tolerations[i] = *toleration.DeepCopy()
		if toleration.Effect == corev1.TaintEffectNoExecute {
			// TolerationSeconds is ignored unless Effect is
			// NoExecute.
			tolerations[i].TolerationSeconds = nil
		}
	}
	sort.Slice(tolerations, func(i, j int) bool {
		return tolerations[i].Key < tolerations[j].Key || tolerations[i].Operator < tolerations[j].Operator || tolerations[i].Value < tolerations[j].Value || tolerations[i].Effect < tolerations[j].Effect
	})
	hashableDaemonSet.Spec.Template.Spec.Tolerations = tolerations
	hashableDaemonSet.Spec.UpdateStrategy = daemonset.Spec.UpdateStrategy
	volumes := make([]corev1.Volume, len(daemonset.Spec.Template.Spec.Volumes))
	for i, vol := range daemonset.Spec.Template.Spec.Volumes {
		volumes[i] = *vol.DeepCopy()
		// 420 is the default value for DefaultMode for ConfigMap and
		// Secret volumes.
		if vol.ConfigMap != nil && vol.ConfigMap.DefaultMode != nil && *vol.ConfigMap.DefaultMode == int32(420) {
			volumes[i].ConfigMap.DefaultMode = nil
		}
		if vol.Secret != nil && vol.Secret.DefaultMode != nil && *vol.Secret.DefaultMode == int32(420) {
			volumes[i].Secret.DefaultMode = nil
		}
	}
	sort.Slice(volumes, func(i, j int) bool {
		return volumes[i].Name < volumes[j].Name
	})
	hashableDaemonSet.Spec.Template.Spec.Volumes = volumes

	return &hashableDaemonSet
}

// deleteEnvoyDaemonSet deletes an Envoy daemonset.
func (r *reconciler) deleteEnvoyDaemonSet(daemonset *appsv1.DaemonSet) error {
	if err := r.client.Delete(context.TODO(), daemonset); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete Envoy daemonset %s/%s: %v", daemonset.Namespace, daemonset.Name, err)
		}
	} else {
		log.Info("deleted Envoy daemonset", "namespace", daemonset.Namespace, "name", daemonset.Name)
	}
	return nil
}

// currentEnvoyDaemonSet returns the current Envoy daemonset.
func (r *reconciler) currentEnvoyDaemonSet(ic *operatorv1.IngressController) (bool, *appsv1.DaemonSet, error) {
	daemonset := &appsv1.DaemonSet{}
	if err := r.client.Get(context.TODO(), controller.EnvoyDaemonSetName(ic), daemonset); err != nil {
		if errors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, daemonset, nil
}

// createEnvoyDaemonSet creates an Envoy daemonset.
func (r *reconciler) createEnvoyDaemonSet(daemonset *appsv1.DaemonSet) error {
	if err := r.client.Create(context.TODO(), daemonset); err != nil {
		return fmt.Errorf("failed to create Envoy daemonset %s/%s: %v", daemonset.Namespace, daemonset.Name, err)
	}
	log.Info("created Envoy daemonset", "namespace", daemonset.Namespace, "name", daemonset.Name)
	return nil
}

// updateEnvoyDaemonSet updates an Envoy daemonset.
func (r *reconciler) updateEnvoyDaemonSet(current, desired *appsv1.DaemonSet) error {
	changed, updated := daemonsetConfigChanged(current, desired)
	if !changed {
		return nil
	}

	if err := r.client.Update(context.TODO(), updated); err != nil {
		return fmt.Errorf("failed to update Envoy daemonset %s/%s: %v", updated.Namespace, updated.Name, err)
	}
	log.Info("updated Envoy daemonset", "namespace", updated.Namespace, "name", updated.Name)
	return nil
}

// daemonsetConfigChanged checks if current config matches the expected config
// for the ingress controller daemonset and if not returns the updated config.
func daemonsetConfigChanged(current, expected *appsv1.DaemonSet) (bool, *appsv1.DaemonSet) {
	if daemonsetHash(current) == daemonsetHash(expected) {
		return false, nil
	}

	diff := cmp.Diff(current, expected)
	log.Info("current Envoy daemonset differs from expected", "namespace", current.Namespace, "name", current.Name, "diff", diff)

	updated := current.DeepCopy()
	updated.Labels = expected.Labels
	updated.Spec.Template.Labels = expected.Spec.Template.Labels
	// Copy the primary container from current and update its fields
	// selectively.  Copy any sidecars from expected verbatim.
	containers := make([]corev1.Container, len(expected.Spec.Template.Spec.Containers))
	containers[0] = updated.Spec.Template.Spec.Containers[0]
	for i, container := range expected.Spec.Template.Spec.Containers[1:] {
		containers[i+1] = *container.DeepCopy()
	}
	updated.Spec.Template.Spec.Containers = containers
	updated.Spec.Template.Spec.Containers[0].Args = expected.Spec.Template.Spec.Containers[0].Args
	updated.Spec.Template.Spec.Containers[0].Command = expected.Spec.Template.Spec.Containers[0].Command
	updated.Spec.Template.Spec.Containers[0].Env = expected.Spec.Template.Spec.Containers[0].Env
	updated.Spec.Template.Spec.Containers[0].Image = expected.Spec.Template.Spec.Containers[0].Image
	updated.Spec.Template.Spec.Containers[0].ImagePullPolicy = expected.Spec.Template.Spec.Containers[0].ImagePullPolicy
	updated.Spec.Template.Spec.Containers[0].VolumeMounts = expected.Spec.Template.Spec.Containers[0].VolumeMounts
	updated.Spec.Template.Spec.HostNetwork = expected.Spec.Template.Spec.HostNetwork
	updated.Spec.Template.Spec.NodeSelector = expected.Spec.Template.Spec.NodeSelector
	updated.Spec.Template.Spec.Tolerations = expected.Spec.Template.Spec.Tolerations
	volumes := make([]corev1.Volume, len(expected.Spec.Template.Spec.Volumes))
	for i, vol := range expected.Spec.Template.Spec.Volumes {
		volumes[i] = *vol.DeepCopy()
	}
	updated.Spec.Template.Spec.Volumes = volumes
	updated.Spec.UpdateStrategy = expected.Spec.UpdateStrategy

	return true, updated
}
