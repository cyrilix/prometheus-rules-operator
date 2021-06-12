package rules

import (
	"embed"
	monitoringv1alpha1 "github.com/cyrilix/prometheus-rules-operator/api/v1alpha1"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/r3labs/diff/v2"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"testing"
)

//go:embed testData/*.yaml
var sampleRuleFiles embed.FS

var appsGroup pov1.RuleGroup
var resourcesGroup pov1.RuleGroup

func init() {
	rg, err := loadRulesFiles("testData", &sampleRuleFiles)
	if err != nil {
		log.Panicf("unable to load test rules: %v", err)
	}
	for _, grp := range rg {
		switch grp.Name {
		case "kubernetes-apps":
			appsGroup = grp
		case "kubernetes-resources":
			resourcesGroup = grp
		}
	}
}

func Test_loadRules(t *testing.T) {
	tests := []struct {
		name           string
		wantGroupCount int
		wantErr        bool
	}{
		{"load-all-rules", 16, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadRules()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantGroupCount {
				t.Errorf("loadRules() got = %v, want %v", len(got), tt.wantGroupCount)
			}
		})
	}
}

func Test_filterAndPatchGroup(t *testing.T) {
	allGroups := []pov1.RuleGroup{appsGroup, resourcesGroup}
	type args struct {
		groups  []pov1.RuleGroup
		patches []monitoringv1alpha1.GroupPatch
	}
	tests := []struct {
		name string
		args args
		want []pov1.RuleGroup
	}{
		{
			"no-transformation",
			args{
				allGroups,
				[]monitoringv1alpha1.GroupPatch{},
			},
			allGroups,
		},
		{
			"exclude-apps-group",
			args{
				allGroups,
				[]monitoringv1alpha1.GroupPatch{
					{
						Name:      "kubernetes-apps",
						SkipGroup: true,
					},
				},
			},
			[]pov1.RuleGroup{resourcesGroup},
		},
		{
			"exclude-rule",
			args{
				[]pov1.RuleGroup{resourcesGroup},
				[]monitoringv1alpha1.GroupPatch{
					{
						Name: "kubernetes-resources",
						RulePatches: []monitoringv1alpha1.RulePatch{
							{
								Name:     "KubeCPUOvercommit",
								SkipRule: true,
							},
						},
					},
				},
			},
			[]pov1.RuleGroup{
				{
					Name:  "kubernetes-resources",
					Rules: []pov1.Rule{},
				},
			},
		},
		{
			"exclude-one-rule",
			args{
				[]pov1.RuleGroup{appsGroup},
				[]monitoringv1alpha1.GroupPatch{
					{
						Name: "kubernetes-apps",
						RulePatches: []monitoringv1alpha1.RulePatch{
							{
								Name:     "KubePodCrashLooping",
								SkipRule: true,
							},
						},
					},
				},
			},
			[]pov1.RuleGroup{
				{
					Name: "kubernetes-apps",
					Rules: []pov1.Rule{
						{
							Alert: "KubePodNotReady",
							Annotations: map[string]string{
								"description": "Pod {{ $labels.namespace }}/{{ $labels.pod }} has been in a non-ready state for longer than 15 minutes.",
								"runbook_url": "https://github.com/prometheus-operator/kube-prometheus/wiki/kubepodnotready",
								"summary":     "Pod has been in a non-ready state for more than 15 minutes.",
							},
							Expr:   intstr.Parse("sum by (namespace, pod) (\n  max by(namespace, pod) (\n    kube_pod_status_phase{job=\"kube-state-metrics\", phase=~\"Pending|Unknown\"}\n  ) * on(namespace, pod) group_left(owner_kind) topk by(namespace, pod) (\n    1, max by(namespace, pod, owner_kind) (kube_pod_owner{owner_kind!=\"Job\"})\n  )\n) > 0\n"),
							For:    "15m",
							Labels: map[string]string{"severity": "warning"},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterAndPatchGroup(tt.args.groups, tt.args.patches); !reflect.DeepEqual(got, tt.want) {
				d, _ := diff.NewDiffer()

				changelog, _ := d.Diff(got, tt.want)
				log.Errorf("diff: %#v", changelog)

				t.Errorf("FilterAndPatchGroup() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
