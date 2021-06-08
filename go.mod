module github.com/cyrilix/prometheus-rules-operator

go 1.16

require (
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.48.1
	github.com/r3labs/diff/v2 v2.13.1
	github.com/sirupsen/logrus v1.6.0
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/yaml v1.2.0
)
