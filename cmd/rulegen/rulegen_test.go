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
	_ "embed"
	"fmt"
	"github.com/andreyvit/diff"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"
)

var (
	//go:embed testData/ruleExample.yaml
	rule1Example string

	//go:embed testData/rawRuleExample.yaml
	rule2Example string

	//go:embed testData/expectedRules.yaml
	expectedRules string
)

func Test_run(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		switch request.RequestURI {
		case "/rules1":
			writer.WriteHeader(http.StatusOK)
			_, err := writer.Write([]byte(rule1Example))
			if err != nil {
				log.Errorf("unable to write content response: %v", err)
			}
		case "/rules2":
			writer.WriteHeader(http.StatusOK)
			_, err := writer.Write([]byte(rule2Example))
			if err != nil {
				log.Errorf("unable to write content response: %v", err)
			}
		default:
			log.Errorf("no rule for uri '%s'", request.RequestURI)
			writer.WriteHeader(http.StatusNotFound)
		}
	}))

	defer server.Close()
	type args struct {
		rulePath         string
		destFile         string
		operatorResource bool
	}
	tests := []struct {
		name          string
		args          args
		expectedRules string
		wantErr       bool
	}{
		{"operator-rule", args{"rules1", "kubernetes-monitoring-rules.yaml", true}, expectedRules, false},
		{"raw-rule", args{"rules2", "rule2.yaml", false}, expectedRules, false},
	}

	for _, tt := range tests {
		resultDir := t.TempDir()
		t.Run(tt.name, func(t *testing.T) {
			destFile := path.Join(resultDir, tt.args.destFile)
			urlRule, err := url.Parse(fmt.Sprintf("%s/%s", server.URL, tt.args.rulePath))
			if err != nil {
				t.Errorf("unable to generate url test: %v", err)
				return
			}

			if err := run(urlRule, destFile, tt.args.operatorResource); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				content, err := ioutil.ReadFile(destFile)
				if err != nil {
					t.Errorf("unable to read generated file content: %v", err)
					return
				}
				result := strings.Replace(string(content), urlRule.String(), "https://url/rule.yaml", -1)
				if result != tt.expectedRules {
					t.Errorf("bad generated file content: %s", diff.LineDiff(expectedRules, result))
				}
			}
		})
	}
}
