/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sigs.k8s.io/yaml"
	"text/template"
)

const generatedFileTemplate = `# DON'T MANUALLY EDIT THIS FILE
# This file is generated from "{{ .importUrl }}", run "go generate ./..." to update it.

{{ .ruleContent }}`

var headerTpl = template.Must(template.New("header").Parse(generatedFileTemplate))

func main() {

	var ruleUrl, destFile string
	var isPrometheusOperatorResource bool
	flag.StringVar(&ruleUrl, "rule-url", "", "Prometheus rule url to import.")
	flag.StringVar(&destFile, "dest-file", ".", "Directory where to write generated co`e.")
	flag.BoolVar(&isPrometheusOperatorResource, "prometheus-operator-resource", false, "If true, manage content as prometheus-opertor rule")
	flag.Parse()

	log.Infof("generate prometheus rules from %v", ruleUrl)
	rule, err := url.Parse(ruleUrl)
	if err != nil {
		log.Panicf("unable to parse rule url '%s': %v", ruleUrl, err)
	}

	if err := run(rule, destFile, isPrometheusOperatorResource); err != nil {
		log.Fatalf("unable to synchronize url rule '%s' into '%s' dir: %v", ruleUrl, destFile, err)
	}

}

func run(urlRule *url.URL, destFile string, isPrometheusOperatorResource bool) error {
	ctx := context.Background()
	client := http.DefaultClient

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlRule.String(), bytes.NewBufferString(""))
	if err != nil {
		return fmt.Errorf("unable to build http request for prometheus rule '%s': %v", urlRule, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to fetch prometheus rule '%s': %v", urlRule, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code for prometheus rule %s: %v - %s", urlRule, resp.StatusCode, resp.Status)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read content for prometheus rule %s: %v", urlRule, err)
	}

	var yamlContent []byte
	if isPrometheusOperatorResource {
		jsonContent, err := yaml.YAMLToJSON(content)
		if err != nil {
			return fmt.Errorf("unable to convert yaml rules '%s' to json: %v", urlRule, err)
		}
		var rules pov1.PrometheusRule
		err = json.Unmarshal(jsonContent, &rules)
		if err != nil {
			return fmt.Errorf("unable to unmarshal prometheus rules '%s' with content '%s': %v", urlRule, jsonContent, err)
		}

		ruleGroups := rules.Spec
		rulesJson, err := json.Marshal(ruleGroups)
		if err != nil {
			return fmt.Errorf("unable to marshal rules to json: %v", err)
		}

		yamlContent, err = yaml.JSONToYAML(rulesJson)
		if err != nil {
			return fmt.Errorf("unable to marshal rules groups to yaml: %v", err)
		}
	} else {
		yamlContent = content
	}

	resultFile, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("unable to create %s file: %v", destFile, err)
	}
	defer func() {
		if err := resultFile.Close(); err != nil {
			log.Errorf("unable to close %s file: %v", destFile, err)
		}
	}()
	err = headerTpl.Execute(resultFile, map[string]string{"importUrl": urlRule.String(), "ruleContent": string(yamlContent)})
	if err != nil {

	}
	return nil
}
