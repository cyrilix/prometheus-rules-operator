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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RuleSpec defines the desired state of Rule
type RuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Rule. Edit rule_types.go to remove/update
	GroupPatch []GroupPatch `json:"groupPatch,omitempty"`
}

// RuleStatus defines the observed state of Rule
type RuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Rule is the Schema for the rules API
type Rule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RuleSpec   `json:"spec,omitempty"`
	Status RuleStatus `json:"status,omitempty"`
}

type GroupName string
type RuleName string

type GroupPatch struct {
	Name        GroupName   `json:"name"`
	SkipGroup   bool        `json:"skip"`
	RulePatches []RulePatch `json:"rule_patches"`
}

func (gp *GroupPatch) HasRulePatches() bool {
	return len(gp.RulePatches) > 0
}

func (gp *GroupPatch) RuleIndex() map[RuleName]*RulePatch {
	idx := make(map[RuleName]*RulePatch, len(gp.RulePatches))
	for _, rp := range gp.RulePatches {
		idx[rp.Name] = &rp
	}
	return idx
}

type RulePatch struct {
	Name     RuleName `json:"name"`
	SkipRule bool     `json:"skip"`
}

//+kubebuilder:object:root=true

// RuleList contains a list of Rule
type RuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rule{}, &RuleList{})
}
