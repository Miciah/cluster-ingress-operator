module github.com/openshift/cluster-ingress-operator

go 1.12

require (
	github.com/Azure/azure-sdk-for-go v35.0.0+incompatible
	github.com/Azure/go-autorest/autorest v0.9.0
	github.com/Azure/go-autorest/autorest/adal v0.5.0
	github.com/aws/aws-sdk-go v1.16.26
	github.com/coreos/bbolt v1.3.1-coreos.6 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/google/go-cmp v0.3.0
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/kevinburke/go-bindata v3.11.0+incompatible
	github.com/openshift/api v3.9.1-0.20191028134408-7e36eed0d19e+incompatible
	github.com/openshift/library-go v0.0.0-20190927184318-c355e2019bb3
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	go.uber.org/zap v1.10.0
	google.golang.org/api v0.6.1-0.20190607001116-5213b8090861
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/yaml.v2 v2.2.4

	// kubernetes-1.16.0
	k8s.io/api v0.17.1
	k8s.io/apimachinery v0.17.1
	k8s.io/apiserver v0.17.1
	k8s.io/client-go v0.17.1
	k8s.io/cloud-provider v0.17.1
	k8s.io/legacy-cloud-providers v0.17.1

	sigs.k8s.io/controller-runtime v0.3.1-0.20191011155846-b2bc3490f2e3
	sigs.k8s.io/controller-tools v0.2.2-0.20190919191502-76a25b63325a
)
