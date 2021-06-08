package rules

import (
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generatePrometheusRules(name string) *pov1.PrometheusRule {
	rules := pov1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: pov1.PrometheusRuleSpec{},
	}

	return &rules
	//&rules
}
