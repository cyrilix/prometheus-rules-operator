package rules

//go:generate go run ../cmd/rulegen -prometheus-operator-resource=true -dest-file ./kubernetes-prometheusRule.yaml -rule-url "https://raw.githubusercontent.com/prometheus-operator/kube-prometheus/master/manifests/kubernetes-prometheusRule.yaml"
//go:generate go run ../cmd/rulegen -prometheus-operator-resource=false -dest-file ./etcd3_alert.rules.yaml -rule-url "https://raw.githubusercontent.com/etcd-io/website/master/content/en/docs/v3.4/op-guide/etcd3_alert.rules.yml"

import (
	"embed"
	"encoding/json"
	"fmt"
	monitoringv1alpha1 "github.com/cyrilix/prometheus-rules-operator/api/v1alpha1"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"log"
	"path"
	"sigs.k8s.io/yaml"
)

//go:embed *.yaml
var ruleFiles embed.FS

func MustLoadRules() []pov1.RuleGroup {
	rg, err := loadRules()
	if err != nil {
		log.Panicf("unable to load rule groups: %v", err)
	}
	return rg
}

func loadRules() ([]pov1.RuleGroup, error) {
	return loadRulesFiles(".", &ruleFiles)
}

func loadRulesFiles(filePath string, files *embed.FS) ([]pov1.RuleGroup, error) {
	// list files
	result := make([]pov1.RuleGroup, 0, 10)
	// for each, load rule
	entries, err := files.ReadDir(filePath)
	if err != nil {
		return []pov1.RuleGroup{}, fmt.Errorf("unable to list yaml rules files")
	}

	for _, entry := range entries {
		content, err := files.ReadFile(path.Join(filePath, entry.Name()))
		if err != nil {
			return []pov1.RuleGroup{}, fmt.Errorf("unable to read file '%v': %v", entry.Name(), err)
		}

		jsonContent, err := yaml.YAMLToJSON(content)
		if err != nil {
			return []pov1.RuleGroup{}, fmt.Errorf("unable to convert yaml rules '%s' to json: %v", entry.Name(), err)
		}
		var rulesSpec pov1.PrometheusRuleSpec
		err = json.Unmarshal(jsonContent, &rulesSpec)
		if err != nil {
			return []pov1.RuleGroup{}, fmt.Errorf("unable to unmarshal prometheus rules '%s' with content '%s': %v", entry.Name(), jsonContent, err)
		}
		result = append(result, rulesSpec.Groups...)
	}

	return result, nil
}

func FilterAndPatchGroup(groups []pov1.RuleGroup, patches []monitoringv1alpha1.GroupPatch) []pov1.RuleGroup {
	result := make([]pov1.RuleGroup, 0, len(groups))

	idxPatches := make(map[monitoringv1alpha1.GroupName]*monitoringv1alpha1.GroupPatch, len(patches))
	for _, pg := range patches {
		idxPatches[pg.Name] = &pg
	}

	for _, grp := range groups {

		if _, ok := idxPatches[monitoringv1alpha1.GroupName(grp.Name)]; !ok {
			// no patch, keep GroupRule
			result = append(result, grp)
			continue
		}

		pg := idxPatches[monitoringv1alpha1.GroupName(grp.Name)]
		if pg.SkipGroup {
			// ignore this group
			continue
		}

		if !pg.HasRulePatches() {
			result = append(result, grp)
			continue
		}

		// Patch Rules
		rulesIdx := pg.RuleIndex()
		patchedRules := make([]pov1.Rule, 0, len(grp.Rules))
		for _, r := range grp.Rules {
			p, ok := rulesIdx[monitoringv1alpha1.RuleName(r.Alert)]
			if !ok {
				// no patch, add rule as it
				patchedRules = append(patchedRules, r)
				continue
			}

			if !p.SkipRule {
				// no patch, add rule as it
				patchedRules = append(patchedRules, r)
			}
		}
		grp.Rules = patchedRules
		result = append(result, grp)
	}
	return result
}
