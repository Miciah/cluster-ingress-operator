package ingress

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"text/template"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-ingress-operator/pkg/operator/controller"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var envoyJsonTemplate = template.Must(template.New("envoy.json").Parse(`{
  "static_resources": {
    "clusters": [
      {
        "name": "contour",
        "alt_stat_name": "{{.Namespace}}_contour_8001",
        "type": "STRICT_DNS",
        "connect_timeout": "5s",
        "load_assignment": {
          "cluster_name": "contour",
          "endpoints": [
            {
              "lb_endpoints": [
                {
                  "endpoint": {
                    "address": {
                      "socket_address": {
                        "address": "%s",
                        "port_value": 8001
                      }
                    }
                  }
                }
              ]
            }
          ]
        },
        "circuit_breakers": {
          "thresholds": [
            {
              "priority": "HIGH",
              "max_connections": 100000,
              "max_pending_requests": 100000,
              "max_requests": 60000000,
              "max_retries": 50
            },
            {
              "max_connections": 100000,
              "max_pending_requests": 100000,
              "max_requests": 60000000,
              "max_retries": 50
            }
          ]
        },
        "http2_protocol_options": {},
        "transport_socket": {
          "name": "envoy.transport_sockets.tls",
          "typed_config": {
            "@type": "type.googleapis.com/envoy.api.v2.auth.UpstreamTlsContext",
            "common_tls_context": {
              "tls_certificates": [
                {
                  "certificate_chain": {
                    "filename": "/certs/tls.crt"
                  },
                  "private_key": {
                    "filename": "/certs/tls.key"
                  }
                }
              ],
              "validation_context": {
                "trusted_ca": {
                  "filename": "/ca/cacert.pem"
                },
                "verify_subject_alt_name": [
                  "{{.ContourServiceName}}"
                ]
              }
            }
          }
        },
        "upstream_connection_options": {
          "tcp_keepalive": {
            "keepalive_probes": 3,
            "keepalive_time": 30,
            "keepalive_interval": 5
          }
        }
      },
      {
        "name": "service-stats",
        "alt_stat_name": "{{.Namespace}}_service-stats_9001",
        "type": "LOGICAL_DNS",
        "connect_timeout": "0.250s",
        "load_assignment": {
          "cluster_name": "service-stats",
          "endpoints": [
            {
              "lb_endpoints": [
                {
                  "endpoint": {
                    "address": {
                      "socket_address": {
                        "address": "127.0.0.1",
                        "port_value": 9001
                      }
                    }
                  }
                }
              ]
            }
          ]
        }
      }
    ]
  },
  "dynamic_resources": {
    "lds_config": {
      "api_config_source": {
        "api_type": "GRPC",
        "grpc_services": [
          {
            "envoy_grpc": {
              "cluster_name": "contour"
            }
          }
        ]
      }
    },
    "cds_config": {
      "api_config_source": {
        "api_type": "GRPC",
        "grpc_services": [
          {
            "envoy_grpc": {
              "cluster_name": "contour"
            }
          }
        ]
      }
    }
  },
  "admin": {
    "access_log_path": "/dev/null",
    "address": {
      "socket_address": {
        "address": "127.0.0.1",
        "port_value": 9001
      }
    }
  }
}
`))

// ensureEnvoyConfigMap ensures the Envoy configmap exists for a
// given ingresscontroller if it uses Envoy.  Returns a Boolean
// indicating whether the configmap exists, the configmap if it does
// exist, and an error value.
func (r *reconciler) ensureEnvoyConfigMap(ic *operatorv1.IngressController, deploymentRef metav1.OwnerReference) (bool, *corev1.ConfigMap, error) {
	want, desired, err := desiredEnvoyConfigMap(ic, deploymentRef)
	if err != nil {
		return false, nil, err
	}

	have, current, err := r.currentEnvoyConfigMap(ic)
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
		if updated, err := r.updateEnvoyConfigMap(current, desired); err != nil {
			return true, nil, fmt.Errorf("failed to update configmap: %v", err)
		} else if updated {
			log.Info("updated configmap", "configmap", desired)
		}
	}

	return r.currentEnvoyConfigMap(ic)
}

// desiredEnvoyConfigMap returns the desired Envoy configmap.  Returns a
// Boolean indicating whether a configmap is desired, as well as the configmap
// if one is desired.
func desiredEnvoyConfigMap(ic *operatorv1.IngressController, deploymentRef metav1.OwnerReference) (bool, *corev1.ConfigMap, error) {
	if !UseContour(ic) {
		return false, nil, nil
	}

	envoyConfigMapName := controller.EnvoyConfigMapName(ic)
	contourServiceName := controller.ContourServiceName(ic)
	envoyJsonParameters := struct {
		Namespace          string
		ContourServiceName string
	}{
		Namespace:          envoyConfigMapName.Namespace,
		ContourServiceName: contourServiceName.Name,
	}
	envoyJson := new(bytes.Buffer)
	if err := envoyJsonTemplate.Execute(envoyJson, envoyJsonParameters); err != nil {
		return false, nil, err
	}
	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      envoyConfigMapName.Name,
			Namespace: envoyConfigMapName.Namespace,
		},
		Data: map[string]string{
			"envoy.json": envoyJson.String(),
		},
	}
	cm.SetOwnerReferences([]metav1.OwnerReference{deploymentRef})

	return true, &cm, nil
}

// currentEnvoyConfigMap returns the current Envoy configmap.  Returns a
// Boolean indicating whether the configmap existed, the configmap if it did
// exist, and an error value.
func (r *reconciler) currentEnvoyConfigMap(ic *operatorv1.IngressController) (bool, *corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	if err := r.client.Get(context.TODO(), controller.EnvoyConfigMapName(ic), cm); err != nil {
		if errors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, cm, nil
}

// updateEnvoyConfigMap updates a configmap.  Returns a Boolean indicating
// whether the configmap was updated, and an error value.
func (r *reconciler) updateEnvoyConfigMap(current, desired *corev1.ConfigMap) (bool, error) {
	if envoyConfigmapsEqual(current, desired) {
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

// envoyConfigmapsEqual compares two Envoy configmaps.  Returns true if the
// configmaps should be considered equal for the purpose of determining whether
// an update is necessary, false otherwise
func envoyConfigmapsEqual(a, b *corev1.ConfigMap) bool {
	if !reflect.DeepEqual(a.Data, b.Data) {
		return false
	}
	return true
}
