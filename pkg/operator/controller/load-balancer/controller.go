package loadbalancer

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"

	cloudprovider "k8s.io/cloud-provider"
	_ "k8s.io/legacy-cloud-providers/aws"
	_ "k8s.io/legacy-cloud-providers/azure"
	_ "k8s.io/legacy-cloud-providers/gce"
	_ "k8s.io/legacy-cloud-providers/openstack"
	_ "k8s.io/legacy-cloud-providers/vsphere"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"

	"k8s.io/client-go/tools/record"

	logf "github.com/openshift/cluster-ingress-operator/pkg/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"

	configv1 "github.com/openshift/api/config/v1"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimecontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	controllerName = "loadbalancer_controller"

	// cloudCredentialsSecretName is the name of the secret in the
	// operator's namespace that will hold the credentials that the operator
	// will use to authenticate with the cloud API.
	cloudCredentialsSecretName = "cloud-credentials"

	// needLoadBalancerLabel indicates that a so-labeled service needs a
	// load balancer.
	needLoadBalancerLabel = "ingresscontroller.operator.openshift.io/need-load-balancer"
)

var (
	log = logf.Logger.WithName(controllerName)

	// needLoadBalancerSelector is a label selector for
	// needLoadBalancerLabel.
	needLoadBalancerSelector = labels.NewSelector().Add(MustNewRequirement(needLoadBalancerLabel, selection.Exists, nil))
)

func MustNewRequirement(key string, op selection.Operator, vals []string) labels.Requirement {
	result, err := labels.NewRequirement(key, op, vals)
	if err != nil {
		panic(err)
	}
	if result == nil {
		panic("NewRequirement returned nil")
	}

	return *result
}

func New(mgr manager.Manager, operatorNamespace, operatorReleaseVersion string) (runtimecontroller.Controller, error) {
	cl := mgr.GetClient()

	infrastructure := &configv1.Infrastructure{}
	if err := cl.Get(context.TODO(), types.NamespacedName{Name: "cluster"}, infrastructure); err != nil {
		return nil, fmt.Errorf("failed to get infrastructure 'cluster': %v", err)
	}

	var balancer cloudprovider.LoadBalancer
	if haveCloudProvider, cloudProviderName := getCloudProvider(infrastructure); !haveCloudProvider {
		log.Info("platform does not have a cloud provider implementation", "platform", infrastructure.Status.PlatformStatus.Type)
		return nil, nil
	} else {
		log.Info("initializing cloud provider", "platform", infrastructure.Status.PlatformStatus.Type, "cloud provider", cloudProviderName)

		var configFileName, configContent string
		var needConfig bool

		switch infrastructure.Status.PlatformStatus.Type {
		case configv1.AWSPlatformType:
			needConfig = true
			configContent = "[Global]\n"

			clusterId := infrastructure.Status.InfrastructureName
			configContent = configContent + "kubernetesClusterId=" + clusterId + "\n"

			// We don't have the zone, but we know that it is the
			// same as the region with a 1-character suffix, and the
			// cloud provider will remove that suffix from the zone
			// in order to determine the region (which is what the
			// cloud provider code really needs).
			region := infrastructure.Status.PlatformStatus.AWS.Region
			zone := region + "x"
			configContent = configContent + "zone=" + zone + "\n"
			configContent = configContent + "disableStrictZoneCheck=true\n"

			credsSecret := &corev1.Secret{}
			credsSecretName := types.NamespacedName{Namespace: operatorNamespace, Name: cloudCredentialsSecretName}
			err := cl.Get(context.TODO(), credsSecretName, credsSecret)
			if err != nil {
				return nil, fmt.Errorf("failed to get cloud credentials from secret %v: %v", credsSecretName, err)
			}

			creds := credentials.NewStaticCredentials(string(credsSecret.Data["aws_access_key_id"]), string(credsSecret.Data["aws_secret_access_key"]), "")
			sess, err := session.NewSessionWithOptions(session.Options{
				Config: aws.Config{
					Credentials: creds,
				},
				SharedConfigState: session.SharedConfigEnable,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create AWS session: %v", err)
			}

			sess.Handlers.Build.PushBackNamed(request.NamedHandler{
				Name: "openshift.io/ingress-operator",
				Fn:   request.MakeAddToUserAgentHandler("openshift.io ingress-operator", operatorReleaseVersion),
			})

			iamClient := iam.New(sess)
			getUserOutput, err := iamClient.GetUser(&iam.GetUserInput{})
			if err != nil {
				return nil, fmt.Errorf("failed to get IAM user: %v", err)
			}

			user := getUserOutput.User
			arn := *user.Arn
			configContent = configContent + "roleArn=" + arn + "\n"

			ec2Client := ec2.New(sess, aws.NewConfig().WithRegion(region))
			tag := "tag:kubernetes.io/cluster/" + clusterId
			owned := "owned"
			describeVpcsInput := ec2.DescribeVpcsInput{
				Filters: []*ec2.Filter{{Name: &tag, Values: []*string{&owned}}},
			}
			describeVpcsOutput, err := ec2Client.DescribeVpcs(&describeVpcsInput)
			if err != nil {
				return nil, fmt.Errorf("failed to get VPC: %v", err)
			}
			if len(describeVpcsOutput.Vpcs) != 1 {
				return nil, fmt.Errorf("got unexpected number of VPCs: %d\n%#v", len(describeVpcsOutput.Vpcs), describeVpcsOutput.Vpcs)
			}

			vpc := describeVpcsOutput.Vpcs[0]
			vpcId := *vpc.VpcId
			configContent = configContent + "vpc=" + vpcId + "\n"

			os.Setenv("AWS_ACCESS_KEY_ID", string(credsSecret.Data["aws_access_key_id"]))
			os.Setenv("AWS_SECRET_ACCESS_KEY", string(credsSecret.Data["aws_secret_access_key"]))
		}

		if needConfig {
			configFile, err := ioutil.TempFile("", "cloud-config")
			if err != nil {
				return nil, fmt.Errorf("failed to create cloud-provider config file: %v", err)
			}
			configFileName = configFile.Name()
			defer os.Remove(configFileName)
			if _, err := configFile.Write([]byte(configContent)); err != nil {
				return nil, fmt.Errorf("failed to write cloud-provider config file: %v", err)
			}
			if err := configFile.Close(); err != nil {
				return nil, fmt.Errorf("failed to close cloud-provider config file: %v", err)
			}
			log.Info("wrote config file for cloud provider", "name", configFileName, "content", configContent)
		}

		if cloud, err := cloudprovider.InitCloudProvider(cloudProviderName, configFileName); err != nil {
			return nil, fmt.Errorf("failed to get initialize the cloud provider: %v", err)
		} else if lb, ok := cloud.LoadBalancer(); !ok {
			log.Info("cloud provider does not implement load balancer support", "cloud provider", cloudProviderName)
			return nil, nil
		} else {
			balancer = lb
			log.Info("initialized cloud provider")
		}
	}

	reconciler := &reconciler{
		balancer:    balancer,
		cache:       mgr.GetCache(),
		client:      cl,
		clusterName: infrastructure.Status.InfrastructureName,
		recorder:    mgr.GetEventRecorderFor(controllerName),
	}
	c, err := runtimecontroller.New(controllerName, mgr, runtimecontroller.Options{Reconciler: reconciler})
	if err != nil {
		return nil, err
	}
	if err := c.Watch(
		&source.Kind{Type: &corev1.Node{}},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(reconciler.allServices),
		},
	); err != nil {
		return nil, err
	}
	if err := c.Watch(
		&source.Kind{Type: &corev1.Service{}},
		&handler.EnqueueRequestForObject{},
		predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool { return serviceNeedsLoadBalancer(e.Meta) },
			DeleteFunc: func(e event.DeleteEvent) bool { return serviceNeedsLoadBalancer(e.Meta) },
			UpdateFunc: func(e event.UpdateEvent) bool {
				return serviceNeedsLoadBalancer(e.MetaOld) || serviceNeedsLoadBalancer(e.MetaNew)
			},
			GenericFunc: func(e event.GenericEvent) bool { return serviceNeedsLoadBalancer(e.Meta) },
		},
	); err != nil {
		return nil, err
	}
	return c, nil
}

type reconciler struct {
	balancer    cloudprovider.LoadBalancer
	cache       cache.Cache
	client      client.Client
	clusterName string
	recorder    record.EventRecorder
}

// platformNameToCloudProvider maps platform type to cloud provider.
var platformNameToCloudProvider = map[configv1.PlatformType]string{
	configv1.AWSPlatformType:       "aws",
	configv1.AzurePlatformType:     "azure",
	configv1.VSpherePlatformType:   "vsphere",
	configv1.GCPPlatformType:       "gce",
	configv1.OpenStackPlatformType: "openstack",
}

// getCloudProvider returns a Boolean value indicating whether the cluster uses
// a cloud provider and, if so, a string indicating the cloud provider for the
// provided platform type.
func getCloudProvider(infraConfig *configv1.Infrastructure) (bool, string) {
	var (
		haveCloudProvider bool
		cloudProviderName string
	)

	if status := infraConfig.Status.PlatformStatus; status != nil {
		cloudProviderName, haveCloudProvider = platformNameToCloudProvider[status.Type]
	}

	return haveCloudProvider, cloudProviderName
}

// allServices returns a slice with a reconcile.Request for each service that
// should have a load balancer.
func (r *reconciler) allServices(o handler.MapObject) []reconcile.Request {
	requests := []reconcile.Request{}

	services := &corev1.ServiceList{}
	mls := client.MatchingLabelsSelector{Selector: needLoadBalancerSelector}
	if err := r.cache.List(context.Background(), services, mls); err != nil {
		log.Error(err, "failed to list services", "related", o.Meta.GetSelfLink())
		return requests
	}

	for _, service := range services.Items {
		log.Info("queueing service", "namespace", service.Namespace, "name", service.Name, "related", o.Meta.GetSelfLink())
		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: service.Namespace,
				Name:      service.Name,
			},
		}
		requests = append(requests, request)
	}
	return requests
}

// serviceNeedsLoadBalancer returns a Boolean value indicating whether the
// labels in the provided object metadata include needLoadBalancerLabel.
func serviceNeedsLoadBalancer(meta metav1.Object) bool {
	return needLoadBalancerSelector.Matches(labels.Set(meta.GetLabels()))
}

// Reconcile handles a reconcile request for a resource, which should be a
// service.
func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconciling", "request", request)

	service := &corev1.Service{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, service); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to get service %v: %v", request.NamespacedName, err)
	}

	if service.DeletionTimestamp != nil {
		if err := r.deleteLoadBalancer(service); err != nil {
			log.Error(err, "failed to delete load balancer for service; will retry", "service", service)
			return reconcile.Result{RequeueAfter: 15 * time.Second}, nil
		}

		return reconcile.Result{}, nil
	}

	nodes := &corev1.NodeList{}
	mf := client.MatchingFields(fields.Set{"spec.unschedulable": "false"})
	// TODO: Use a cache or informer.
	if err := r.client.List(context.Background(), nodes, mf); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to list nodes: %v", err)
	}

	readyNodes := []*corev1.Node{}
	for i, node := range nodes.Items {
		if len(node.Status.Conditions) == 0 {
			continue
		}
		ready := false
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				ready = true
				break
			}
		}
		if ready {
			readyNodes = append(readyNodes, &nodes.Items[i])
		}
	}

	if err := r.ensureLoadBalancer(service, readyNodes); err != nil {
		log.Error(err, "failed to ensure load balancer for service; will retry", "service", service)
		return reconcile.Result{RequeueAfter: 15 * time.Second}, nil
	}
	return reconcile.Result{}, nil
}

// ensureLoadBalancer creates load balancer for the given service and updates
// the service's status.
func (r *reconciler) ensureLoadBalancer(service *corev1.Service, nodes []*corev1.Node) error {
	status, err := r.balancer.EnsureLoadBalancer(context.TODO(), r.clusterName, service, nodes)
	if err != nil {
		return fmt.Errorf("failed to ensure load balancer for service %s/%s: %v", service.Namespace, service.Name, err)
	}
	if status == nil {
		return fmt.Errorf("cloud provider returned nil status")
	}

	updated := service.DeepCopy()
	updated.Status.LoadBalancer = *status
	if err := r.client.Status().Update(context.TODO(), updated); err != nil {
		return fmt.Errorf("failed to update %s/%s: %v", service.Namespace, service.Name, err)
	}

	log.Info("ensured load balancer for service", "namespace", service.Namespace, "name", service.Name)

	return nil
}

// deleteLoadBalancer deletes any load-balancer associated with the
// given service.
func (r *reconciler) deleteLoadBalancer(service *corev1.Service) error {
	_, exists, err := r.balancer.GetLoadBalancer(context.TODO(), r.clusterName, service)
	if err != nil {
		return fmt.Errorf("failed to get load balancer: %v", err)
	} else if !exists {
		log.Info("no load balancer exists for service", "namespace", service.Namespace, "name", service.Name)
		return nil
	}

	if err := r.balancer.EnsureLoadBalancerDeleted(context.TODO(), r.clusterName, service); err != nil {
		return fmt.Errorf("failed to delete load balancer: %v", err)
	}

	log.Info("deleted load balancer for service", "namespace", service.Namespace, "name", service.Name)

	return nil
}
