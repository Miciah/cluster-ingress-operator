package ingress

import (
	"context"
	"fmt"
	"hash"
	"hash/fnv"
	"path/filepath"
	"sort"

	"github.com/davecgh/go-spew/spew"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-ingress-operator/pkg/manifests"
	"github.com/openshift/cluster-ingress-operator/pkg/operator/controller"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"

	configv1 "github.com/openshift/api/config/v1"
)

// ensureRouterDeployment ensures the router deployment exists for a given
// ingresscontroller.
func (r *reconciler) ensureRouterDeployment(ci *operatorv1.IngressController, infraConfig *configv1.Infrastructure) (*appsv1.Deployment, error) {
	desired, err := desiredRouterDeployment(ci, r.Config.IngressControllerImage, infraConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build router deployment: %v", err)
	}
	current, err := r.currentRouterDeployment(ci)
	if err != nil {
		return nil, err
	}
	switch {
	case desired != nil && current == nil:
		if err := r.createRouterDeployment(desired); err != nil {
			return nil, err
		}
	case desired != nil && current != nil:
		if err := r.updateRouterDeployment(current, desired); err != nil {
			return nil, err
		}
	}
	return r.currentRouterDeployment(ci)
}

// ensureRouterDeleted ensures that any router resources associated with the
// ingresscontroller are deleted.
func (r *reconciler) ensureRouterDeleted(ci *operatorv1.IngressController) error {
	deployment := &appsv1.Deployment{}
	name := controller.RouterDeploymentName(ci)
	deployment.Name = name.Name
	deployment.Namespace = name.Namespace
	if err := r.client.Delete(context.TODO(), deployment); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

// desiredRouterDeployment returns the desired router deployment.
func desiredRouterDeployment(ci *operatorv1.IngressController, ingressControllerImage string, infraConfig *configv1.Infrastructure) (*appsv1.Deployment, error) {
	deployment := manifests.RouterDeployment()
	name := controller.RouterDeploymentName(ci)
	deployment.Name = name.Name
	deployment.Namespace = name.Namespace

	// Ensure the deployment adopts only its own pods.
	deployment.Spec.Selector = controller.IngressControllerDeploymentPodSelector(ci)
	deployment.Spec.Template.Labels = controller.IngressControllerDeploymentPodSelector(ci).MatchLabels

	needDeploymentHash := false
	switch ci.Status.EndpointPublishingStrategy.Type {
	case operatorv1.HostNetworkStrategyType:
		// Pod replicas for ingress controllers that use the host
		// network cannot be colocated because replicas on the same node
		// would conflict with each other by trying to bind the same
		// ports.  The scheduler avoids scheduling multiple pods that
		// use host networking and specify the same port to the same
		// node.  No affinity policy is required when using HostNetwork.

		// Typically, an ingress controller will be scaled with replicas
		// set equal to the node pool size, in which case, using surge
		// for rolling updates would fail to create new replicas (in the
		// absence of node auto-scaling).  Thus, when using HostNetwork,
		// we set max unavailable to 25% and surge to 0.
		pointerTo := func(ios intstr.IntOrString) *intstr.IntOrString { return &ios }
		deployment.Spec.Strategy = appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxUnavailable: pointerTo(intstr.FromString("25%")),
				MaxSurge:       pointerTo(intstr.FromInt(0)),
			},
		}
	case operatorv1.PrivateStrategyType, operatorv1.LoadBalancerServiceStrategyType:
		// During an upgrade, we want each new pod replica to be
		// colocated with an old replica in order to ensure that a node
		// that had local endpoints at the start of an upgrade continues
		// to have local endpoints for the ingress controller during and
		// at the completion of the upgrade.  To this end, we add a
		// label with a hash of the deployment, using which we can
		// select replicas of the same generation (or select replicas
		// that are *not* of the same generation), configure affinity to
		// colocate replicas of different generations of the same
		// ingress controller, and configure anti-affinity to prevent
		// colocation of replicas of the same generation of the same
		// ingress controller.
		needDeploymentHash = true
		deployment.Spec.Template.Spec.Affinity = &corev1.Affinity{
			PodAffinity: &corev1.PodAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						Weight: int32(100),
						PodAffinityTerm: corev1.PodAffinityTerm{
							TopologyKey: "kubernetes.io/hostname",
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      controller.ControllerDeploymentLabel,
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{controller.IngressControllerDeploymentLabel(ci)},
									},
									{
										Key:      controller.ControllerDeploymentHashLabel,
										Operator: metav1.LabelSelectorOpNotIn,
										// Values is set at the end of this function.
									},
								},
							},
						},
					},
				},
			},
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						TopologyKey: "kubernetes.io/hostname",
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      controller.ControllerDeploymentLabel,
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{controller.IngressControllerDeploymentLabel(ci)},
								},
								{
									Key:      controller.ControllerDeploymentHashLabel,
									Operator: metav1.LabelSelectorOpIn,
									// Values is set at the end of this function.
								},
							},
						},
					},
				},
			},
		}

		// Avoid going from one replica to zero replicas on any given node.
		pointerTo := func(ios intstr.IntOrString) *intstr.IntOrString { return &ios }
		deployment.Spec.Strategy = appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxUnavailable: pointerTo(intstr.FromInt(0)),
				MaxSurge:       pointerTo(intstr.FromInt(1)),
			},
		}
	}

	statsSecretName := fmt.Sprintf("router-stats-%s", ci.Name)
	env := []corev1.EnvVar{
		{Name: "ROUTER_SERVICE_NAME", Value: ci.Name},
		{Name: "STATS_USERNAME", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: statsSecretName,
				},
				Key: "statsUsername",
			},
		}},
		{Name: "STATS_PASSWORD", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: statsSecretName,
				},
				Key: "statsPassword",
			},
		}},
	}

	// Enable prometheus metrics
	certsSecretName := fmt.Sprintf("router-metrics-certs-%s", ci.Name)
	certsVolumeName := "metrics-certs"
	certsVolumeMountPath := "/etc/pki/tls/metrics-certs"

	volume := corev1.Volume{
		Name: certsVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: certsSecretName,
			},
		},
	}
	volumeMount := corev1.VolumeMount{
		Name:      certsVolumeName,
		MountPath: certsVolumeMountPath,
		ReadOnly:  true,
	}

	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMount)

	env = append(env, corev1.EnvVar{Name: "ROUTER_METRICS_TYPE", Value: "haproxy"})
	env = append(env, corev1.EnvVar{Name: "ROUTER_METRICS_TLS_CERT_FILE", Value: filepath.Join(certsVolumeMountPath, "tls.crt")})
	env = append(env, corev1.EnvVar{Name: "ROUTER_METRICS_TLS_KEY_FILE", Value: filepath.Join(certsVolumeMountPath, "tls.key")})

	if len(ci.Status.Domain) > 0 {
		env = append(env, corev1.EnvVar{Name: "ROUTER_CANONICAL_HOSTNAME", Value: ci.Status.Domain})
	}

	if ci.Status.EndpointPublishingStrategy.Type == operatorv1.LoadBalancerServiceStrategyType {
		// For now, check if we are on AWS. This can really be done for
		// for any external [cloud] LBs that support the proxy protocol.
		if infraConfig.Status.Platform == configv1.AWSPlatformType {
			env = append(env, corev1.EnvVar{Name: "ROUTER_USE_PROXY_PROTOCOL", Value: "true"})
		}
	}

	env = append(env, corev1.EnvVar{Name: "ROUTER_THREADS", Value: "4"})

	nodeSelector := map[string]string{
		"kubernetes.io/os":               "linux",
		"node-role.kubernetes.io/worker": "",
	}
	if ci.Spec.NodePlacement != nil {
		if ci.Spec.NodePlacement.NodeSelector != nil {
			var err error
			nodeSelector, err = metav1.LabelSelectorAsMap(ci.Spec.NodePlacement.NodeSelector)
			if err != nil {
				return nil, fmt.Errorf("ingresscontroller %q has invalid spec.nodePlacement.nodeSelector: %v",
					ci.Name, err)
			}
		}
		if ci.Spec.NodePlacement.Tolerations != nil {
			deployment.Spec.Template.Spec.Tolerations = ci.Spec.NodePlacement.Tolerations
		}
	}
	deployment.Spec.Template.Spec.NodeSelector = nodeSelector

	if ci.Spec.NamespaceSelector != nil {
		namespaceSelector, err := metav1.LabelSelectorAsSelector(ci.Spec.NamespaceSelector)
		if err != nil {
			return nil, fmt.Errorf("ingresscontroller %q has invalid spec.namespaceSelector: %v",
				ci.Name, err)
		}

		env = append(env, corev1.EnvVar{
			Name:  "NAMESPACE_LABELS",
			Value: namespaceSelector.String(),
		})
	}

	var desiredReplicas int32 = 2
	if ci.Spec.Replicas != nil {
		desiredReplicas = *ci.Spec.Replicas
	}
	deployment.Spec.Replicas = &desiredReplicas

	if ci.Spec.RouteSelector != nil {
		routeSelector, err := metav1.LabelSelectorAsSelector(ci.Spec.RouteSelector)
		if err != nil {
			return nil, fmt.Errorf("ingresscontroller %q has invalid spec.routeSelector: %v", ci.Name, err)
		}
		env = append(env, corev1.EnvVar{Name: "ROUTE_LABELS", Value: routeSelector.String()})
	}

	deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, env...)

	deployment.Spec.Template.Spec.Containers[0].Image = ingressControllerImage

	if ci.Status.EndpointPublishingStrategy.Type == operatorv1.HostNetworkStrategyType {
		// Expose ports 80 and 443 on the host to provide endpoints for
		// the user's HA solution.
		deployment.Spec.Template.Spec.HostNetwork = true

		// With container networking, probes default to using the pod IP
		// address.  With host networking, probes default to using the
		// node IP address.  Using localhost avoids potential routing
		// problems or firewall restrictions.
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.HTTPGet.Host = "localhost"
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.HTTPGet.Host = "localhost"
	}

	// Fill in the default certificate secret name.
	secretName := controller.RouterEffectiveDefaultCertificateSecretName(ci, deployment.Namespace)
	deployment.Spec.Template.Spec.Volumes[0].Secret.SecretName = secretName.Name

	// If the deployment needs a hash for the affinity policy, we must
	// compute it now, after all the other fields have been computed, and
	// inject it into the appropriate fields.
	if needDeploymentHash {
		hash := deploymentHash(deployment)
		deployment.Spec.Template.Labels[controller.ControllerDeploymentHashLabel] = hash
		values := []string{hash}
		deployment.Spec.Template.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchExpressions[1].Values = values
		deployment.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchExpressions[1].Values = values
	}

	return deployment, nil
}

// deploymentHash returns a stringified hash value for the router deployment
// fields that, if changed, should trigger an update.
func deploymentHash(deployment *appsv1.Deployment) string {
	hasher := fnv.New32a()
	deepHashObject(hasher, hashableDeployment(deployment))
	return rand.SafeEncodeString(fmt.Sprint(hasher.Sum32()))
}

// hashableDeployment returns a copy of the given deployment with exactly the
// fields from deployment that should be used for computing its hash copied
// over.  Fields with slice values will be sorted.  Fields that should be
// ignored, or that have explicit values that are equal to their respective
// default values, will be zeroed.
func hashableDeployment(deployment *appsv1.Deployment) *appsv1.Deployment {
	var hashableDeployment appsv1.Deployment

	var replicas *int32
	if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas != int32(1) {
		// 1 is the default value for Replicas.
		replicas = deployment.Spec.Replicas
	}
	hashableDeployment.Spec.Replicas = replicas

	hashableDeployment.Spec.Strategy = deployment.Spec.Strategy

	hashableDeployment.Spec.Template.Spec.NodeSelector = deployment.Spec.Template.Spec.NodeSelector

	volumes := make([]corev1.Volume, len(deployment.Spec.Template.Spec.Volumes))
	for i, vol := range deployment.Spec.Template.Spec.Volumes {
		volumes[i] = *vol.DeepCopy()
		if vol.Secret != nil && vol.Secret.DefaultMode != nil && *vol.Secret.DefaultMode == int32(420) {
			// 420 is the default value for DefaultMode.
			volumes[i].Secret.DefaultMode = nil
		}
	}
	sort.Slice(volumes, func(i, j int) bool {
		return volumes[i].Name < volumes[j].Name
	})
	hashableDeployment.Spec.Template.Spec.Volumes = volumes

	env := deployment.Spec.Template.Spec.Containers[0].Env
	sort.Slice(env, func(i, j int) bool {
		return env[i].Name < env[j].Name
	})
	image := deployment.Spec.Template.Spec.Containers[0].Image
	hashableDeployment.Spec.Template.Spec.Containers = []corev1.Container{
		{Env: env, Image: image},
	}

	tolerations := make([]corev1.Toleration, len(deployment.Spec.Template.Spec.Tolerations))
	for i, toleration := range deployment.Spec.Template.Spec.Tolerations {
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
	hashableDeployment.Spec.Template.Spec.Tolerations = tolerations

	affinity := deployment.Spec.Template.Spec.Affinity.DeepCopy()
	if affinity != nil {
		cmpMatchExpressions := func(a, b metav1.LabelSelectorRequirement) bool {
			if a.Key != b.Key {
				return a.Key < b.Key
			}
			if a.Operator != b.Operator {
				return a.Operator < b.Operator
			}
			for i := range b.Values {
				if i == len(a.Values) {
					return true
				}
				if a.Values[i] != b.Values[i] {
					return a.Values[i] < b.Values[i]
				}
			}
			return false
		}

		if affinity.PodAffinity != nil {
			terms := affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution
			for _, term := range terms {
				labelSelector := term.PodAffinityTerm.LabelSelector
				if labelSelector != nil {
					for i, expr := range labelSelector.MatchExpressions {
						if expr.Key == controller.ControllerDeploymentHashLabel {
							// Hash value should be ignored.
							labelSelector.MatchExpressions[i].Values = nil
						}
					}

				}
				exprs := labelSelector.MatchExpressions
				sort.Slice(exprs, func(i, j int) bool {
					return cmpMatchExpressions(exprs[i], exprs[j])
				})
			}
		}
		if affinity.PodAntiAffinity != nil {
			terms := affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution
			for _, term := range terms {
				if term.LabelSelector != nil {
					for i, expr := range term.LabelSelector.MatchExpressions {
						if expr.Key == controller.ControllerDeploymentHashLabel {
							// Hash value should be ignored.
							term.LabelSelector.MatchExpressions[i].Values = nil
						}
					}

				}
				exprs := term.LabelSelector.MatchExpressions
				sort.Slice(exprs, func(i, j int) bool {
					return cmpMatchExpressions(exprs[i], exprs[j])
				})
			}
		}
	}
	hashableDeployment.Spec.Template.Spec.Affinity = affinity

	return &hashableDeployment
}

// currentRouterDeployment returns the current router deployment.
func (r *reconciler) currentRouterDeployment(ci *operatorv1.IngressController) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	if err := r.client.Get(context.TODO(), controller.RouterDeploymentName(ci), deployment); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return deployment, nil
}

// createRouterDeployment creates a router deployment.
func (r *reconciler) createRouterDeployment(deployment *appsv1.Deployment) error {
	if err := r.client.Create(context.TODO(), deployment); err != nil {
		return fmt.Errorf("failed to create router deployment %s/%s: %v", deployment.Namespace, deployment.Name, err)
	}
	log.Info("created router deployment", "namespace", deployment.Namespace, "name", deployment.Name)
	return nil
}

// updateRouterDeployment updates a router deployment.
func (r *reconciler) updateRouterDeployment(current, desired *appsv1.Deployment) error {
	changed, updated := deploymentConfigChanged(current, desired)
	if !changed {
		return nil
	}

	if err := r.client.Update(context.TODO(), updated); err != nil {
		return fmt.Errorf("failed to update router deployment %s/%s: %v", updated.Namespace, updated.Name, err)
	}
	log.Info("updated router deployment", "namespace", updated.Namespace, "name", updated.Name)
	return nil
}

// deepHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
//
// Copied from github.com/kubernetes/kubernetes/pkg/util/hash/hash.go.
func deepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}

// deploymentConfigChanged checks if current config matches the expected config
// for the ingress controller deployment and if not returns the updated config.
func deploymentConfigChanged(current, expected *appsv1.Deployment) (bool, *appsv1.Deployment) {
	if deploymentHash(current) == deploymentHash(expected) {
		return false, nil
	}

	updated := current.DeepCopy()
	updated.Spec.Template.Labels = expected.Spec.Template.Labels
	updated.Spec.Strategy = expected.Spec.Strategy
	volumes := make([]corev1.Volume, len(expected.Spec.Template.Spec.Volumes))
	for i, vol := range expected.Spec.Template.Spec.Volumes {
		volumes[i] = *vol.DeepCopy()
	}
	updated.Spec.Template.Spec.Volumes = volumes
	updated.Spec.Template.Spec.NodeSelector = expected.Spec.Template.Spec.NodeSelector
	updated.Spec.Template.Spec.Containers[0].Env = expected.Spec.Template.Spec.Containers[0].Env
	updated.Spec.Template.Spec.Containers[0].Image = expected.Spec.Template.Spec.Containers[0].Image
	updated.Spec.Template.Spec.Tolerations = expected.Spec.Template.Spec.Tolerations
	updated.Spec.Template.Spec.Affinity = expected.Spec.Template.Spec.Affinity
	replicas := int32(1)
	if expected.Spec.Replicas != nil {
		replicas = *expected.Spec.Replicas
	}
	updated.Spec.Replicas = &replicas
	return true, updated
}
