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

// UseContourAnnotation is an annotation that indicates that the
// ingresscontroller uses Contour rather than openshift-router.
const UseContourAnnotation = "ingress.operator.openshift.io/use-contour"

// UseContour returns a Boolean value indicating whether the ingresscontroller
// uses Contour rather than openshift-router.
func UseContour(ic *operatorv1.IngressController) bool {
	val, ok := ic.Annotations[UseContourAnnotation]
	return ok && val != "false"
}

// ensureContourDeployment ensures the Contour deployment exists for a
// given ingresscontroller.
func (r *reconciler) ensureContourDeployment(ic *operatorv1.IngressController, infraConfig *configv1.Infrastructure, ingressConfig *configv1.Ingress, apiConfig *configv1.APIServer, networkConfig *configv1.Network) (bool, *appsv1.Deployment, error) {
	want, desired, err := desiredContourDeployment(ic, r.Config.IngressControllerImage, infraConfig, ingressConfig, apiConfig, networkConfig)
	if err != nil {
		return false, nil, err
	}

	have, current, err := r.currentContourDeployment(ic)
	if err != nil {
		return false, nil, err
	}

	switch {
	case !want && !have:
		return false, nil, nil
	case !want && have:
		if err := r.deleteContourDeployment(current); err != nil {
			return false, nil, err
		}
	case want && !have:
		if err := r.createContourDeployment(desired); err != nil {
			return false, nil, err
		}
	case want && have:
		if err := r.updateContourDeployment(current, desired); err != nil {
			return false, nil, err
		}
	}
	return r.currentContourDeployment(ic)
}

// desiredContourDeployment returns the desired Contour deployment.
func desiredContourDeployment(ic *operatorv1.IngressController, ingressControllerImage string, infraConfig *configv1.Infrastructure, ingressConfig *configv1.Ingress, apiConfig *configv1.APIServer, networkConfig *configv1.Network) (bool, *appsv1.Deployment, error) {
	if !UseContour(ic) {
		return false, nil, nil
	}

	if ic.DeletionTimestamp != nil {
		return false, nil, nil
	}

	var desiredReplicas int32 = 2
	if ic.Spec.Replicas != nil {
		desiredReplicas = *ic.Spec.Replicas
	}

	if ic.Spec.NamespaceSelector != nil {
		_, err := metav1.LabelSelectorAsSelector(ic.Spec.NamespaceSelector)
		if err != nil {
			return false, nil, fmt.Errorf("ingresscontroller %q has invalid spec.namespaceSelector: %v", ic.Name, err)
		}

		return false, nil, fmt.Errorf("ingresscontroller %q both uses Contour and specifies spec.namespaceSelector, which is not supported: %v", ic.Name, err)
	}

	if ic.Spec.RouteSelector != nil {
		_, err := metav1.LabelSelectorAsSelector(ic.Spec.RouteSelector)
		if err != nil {
			return false, nil, fmt.Errorf("ingresscontroller %q has invalid spec.routeSelector: %v", ic.Name, err)
		}

		return false, nil, fmt.Errorf("ingresscontroller %q both uses Contour and specifies spec.routeSelector, which is not supported: %v", ic.Name, err)
	}

	proxyFlag := []string{}
	if ic.Status.EndpointPublishingStrategy.Type == operatorv1.LoadBalancerServiceStrategyType {
		if infraConfig.Status.Platform == configv1.AWSPlatformType {
			proxyFlag = append(proxyFlag, "--use-proxy-protocol")
		}
	}

	fourTwenty := int32(420) // = 0644 octal.
	name := controller.ContourDeploymentName(ic)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
			Labels: map[string]string{
				manifests.OwningIngressControllerLabel: ic.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &desiredReplicas,
			Selector: controller.ContourPodSelector(ic),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "8000",
					},
					Labels: controller.ContourPodSelector(ic).MatchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Args: append([]string{
								"serve",
								"--incluster",
								"--use-extensions-v1beta1-ingress",
								"--xds-address=0.0.0.0",
								"--xds-port=8001",
								"--envoy-service-http-port=80",
								"--envoy-service-https-port=443",
								"--contour-cafile=/ca/cacert.pem",
								"--contour-cert-file=/certs/tls.crt",
								"--contour-key-file=/certs/tls.key",
								"--config-path=/config/contour.yaml",
								// Leader election uses the hard-coded namespace
								// "projectcontour".
								"--disable-leader-election",
							}, proxyFlag...),
							Command: []string{"contour"},
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
							},
							Image:           "docker.io/projectcontour/contour:v1.1.0",
							ImagePullPolicy: corev1.PullIfNotPresent,
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt(8000),
									},
								},
							},
							Name: "contour",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(8001),
									Name:          "xds",
									Protocol:      corev1.ProtocolTCP,
								},
								{
									ContainerPort: int32(8000),
									Name:          "debug",
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromInt(8001),
									},
								},
								InitialDelaySeconds: int32(15),
								PeriodSeconds:       int32(10),
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
									Name:      "contour-config",
									MountPath: "/config",
									ReadOnly:  true,
								},
								{
									Name:      "contourcert",
									MountPath: "/certs",
									ReadOnly:  true,
								},
							},
						},
					},
					DNSPolicy:          corev1.DNSClusterFirst,
					ServiceAccountName: "router",
					Volumes: []corev1.Volume{
						{
							Name: "cacert",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									DefaultMode: &fourTwenty,
									Items: []corev1.KeyToPath{
										{
											Key:  "service-ca.crt",
											Path: "cacert.pem",
										},
									},
									LocalObjectReference: corev1.LocalObjectReference{
										Name: controller.ServiceCAConfigMapName().Name,
									},
								},
							},
						},
						{
							Name: "contour-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									DefaultMode: &fourTwenty,
									Items: []corev1.KeyToPath{
										{
											Key:  "contour.yaml",
											Path: "contour.yaml",
										},
									},
									LocalObjectReference: corev1.LocalObjectReference{
										Name: controller.ContourConfigMapName(ic).Name,
									},
								},
							},
						},
						{
							Name: "contourcert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									DefaultMode: &fourTwenty,
									SecretName:  controller.ContourSecretName(ic).Name,
								},
							},
						},
					},
				},
			},
		},
	}

	return true, deployment, nil
}

// contourDeploymentHash returns a stringified hash value for the Contour
// deployment fields that, if changed, should trigger an update.
func contourDeploymentHash(deployment *appsv1.Deployment) string {
	hasher := fnv.New32a()
	deepHashObject(hasher, hashableContourDeployment(deployment))
	return rand.SafeEncodeString(fmt.Sprint(hasher.Sum32()))
}

// hashableContourDeployment returns a copy of the given Contour deployment with
// exactly the fields from deployment that should be used for computing its hash
// copied over.  In particular, these are the fields that
// desiredContourDeployment sets.  Fields with slice values will be sorted.
// Fields that should be ignored, or that have explicit values that are equal to
// their respective default values, will be zeroed.
func hashableContourDeployment(deployment *appsv1.Deployment) *appsv1.Deployment {
	var hashableDeployment appsv1.Deployment

	hashableDeployment.Labels = deployment.Labels
	hashableDeployment.Name = deployment.Name
	hashableDeployment.Namespace = deployment.Namespace
	var replicas *int32
	if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas != int32(1) {
		// 1 is the default value for Replicas.
		replicas = deployment.Spec.Replicas
	}
	hashableDeployment.Spec.Replicas = replicas
	hashableDeployment.Spec.Template.Labels = deployment.Spec.Template.Labels
	delete(hashableDeployment.Spec.Template.Labels, controller.ControllerDeploymentHashLabel)
	containers := make([]corev1.Container, len(deployment.Spec.Template.Spec.Containers))
	for i, container := range deployment.Spec.Template.Spec.Containers {
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
	hashableDeployment.Spec.Template.Spec.Containers = containers
	volumes := make([]corev1.Volume, len(deployment.Spec.Template.Spec.Volumes))
	for i, vol := range deployment.Spec.Template.Spec.Volumes {
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
	hashableDeployment.Spec.Template.Spec.Volumes = volumes

	return &hashableDeployment
}

// currentContourDeployment returns the current Contour deployment.
func (r *reconciler) currentContourDeployment(ic *operatorv1.IngressController) (bool, *appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	if err := r.client.Get(context.TODO(), controller.ContourDeploymentName(ic), deployment); err != nil {
		if errors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, deployment, nil
}

// deleteContourDeployment deletes a Contour deployment.
func (r *reconciler) deleteContourDeployment(deployment *appsv1.Deployment) error {
	if err := r.client.Delete(context.TODO(), deployment); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete Contour deployment %s/%s: %v", deployment.Namespace, deployment.Name, err)
		}
	} else {
		log.Info("deleted Contour deployment", "namespace", deployment.Namespace, "name", deployment.Name)
	}
	return nil
}

// createContourDeployment creates a Contour deployment.
func (r *reconciler) createContourDeployment(deployment *appsv1.Deployment) error {
	if err := r.client.Create(context.TODO(), deployment); err != nil {
		return fmt.Errorf("failed to create Contour deployment %s/%s: %v", deployment.Namespace, deployment.Name, err)
	}
	log.Info("created Contour deployment", "namespace", deployment.Namespace, "name", deployment.Name)
	return nil
}

// updateContourDeployment updates a Contour deployment.
func (r *reconciler) updateContourDeployment(current, desired *appsv1.Deployment) error {
	changed, updated := contourDeploymentConfigChanged(current, desired)
	if !changed {
		return nil
	}

	if err := r.client.Update(context.TODO(), updated); err != nil {
		return fmt.Errorf("failed to update Contour deployment %s/%s: %v", updated.Namespace, updated.Name, err)
	}
	log.Info("updated Contour deployment", "namespace", updated.Namespace, "name", updated.Name)
	return nil
}

// contourDeploymentConfigChanged checks if current config matches the expected
// config for the Contour deployment and if not returns the updated config.
func contourDeploymentConfigChanged(current, expected *appsv1.Deployment) (bool, *appsv1.Deployment) {
	if contourDeploymentHash(current) == contourDeploymentHash(expected) {
		return false, nil
	}

	diff := cmp.Diff(current, expected)
	log.Info("current Contour deployment differs from expected", "namespace", current.Namespace, "name", current.Name, "diff", diff)

	updated := current.DeepCopy()
	updated.Labels = expected.Labels
	replicas := int32(1)
	if expected.Spec.Replicas != nil {
		replicas = *expected.Spec.Replicas
	}
	updated.Spec.Replicas = &replicas
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
	updated.Spec.Template.Spec.Containers[0].Ports = expected.Spec.Template.Spec.Containers[0].Ports
	updated.Spec.Template.Spec.Containers[0].VolumeMounts = expected.Spec.Template.Spec.Containers[0].VolumeMounts
	volumes := make([]corev1.Volume, len(expected.Spec.Template.Spec.Volumes))
	for i, vol := range expected.Spec.Template.Spec.Volumes {
		volumes[i] = *vol.DeepCopy()
	}
	updated.Spec.Template.Spec.Volumes = volumes

	return true, updated
}
