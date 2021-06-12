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

package controllers

import (
	"context"
	"fmt"
	"github.com/cyrilix/prometheus-rules-operator/rules"
	"github.com/go-logr/logr"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	monitoringv1alpha1 "github.com/cyrilix/prometheus-rules-operator/api/v1alpha1"
)

var ruleGroups = rules.MustLoadRules()

// RuleReconciler reconciles a Rule object
type RuleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=monitoring.cyrilix.fr,resources=rules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.cyrilix.fr,resources=rules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitoring.cyrilix.fr,resources=rules/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Rule object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *RuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// your logic here
	patchRules := &monitoringv1alpha1.Rule{}
	err := r.Get(ctx, req.NamespacedName, patchRules)
	if err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		// Owned objects are automatically garbage collected. For additional cleanup logic, use finalizers.
		// Return and don't requeue
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err = r.reconcileRules(ctx, patchRules, logger); err != nil {
		logger.Error(err, "unable to reconcile resources")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.Rule{}).
		Owns(&pov1.PrometheusRule{}).
		Complete(r)
}

func (r *RuleReconciler) reconcileRules(ctx context.Context, cr *monitoringv1alpha1.Rule, logger logr.Logger) error {
	logger.Info("reconcile PrometheusRules")

	pr, err := r.newPrometheusRule(cr)
	if err != nil {
		return fmt.Errorf("unable to build PrometheusRules: %v from cr: %v", cr, err)
	}

	found := pov1.PrometheusRule{}
	err = r.Get(ctx, types.NamespacedName{Namespace: cr.Namespace, Name: cr.Name}, &found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating new PrometheusRule")
		err = r.Create(ctx, pr)
		if err != nil {
			return fmt.Errorf("unable to create PrometheusRule '%v/%v': %v", cr.Namespace, cr.Name, err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("unable to read PrometheusRule '%v/%v': %v", cr.Namespace, cr.Name, err)
	}

	if !reflect.DeepEqual(pr.Spec, found.Spec) {
		found.Spec = pr.Spec
		logger.Info("Update PrometheusRule " + found.Name)
		err := r.Update(ctx, &found)
		if err != nil {
			return fmt.Errorf("unable to update PrometheusRule '%v/%v': %v", found.Namespace, found.Name, err)
		}
	}

	return nil
}

func (r *RuleReconciler) newPrometheusRule(cr *monitoringv1alpha1.Rule) (*pov1.PrometheusRule, error) {
	pr := &pov1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: pov1.PrometheusRuleSpec{
			Groups: rules.FilterAndPatchGroup(ruleGroups, cr.Spec.GroupPatch),
		},
	}
	if err := controllerutil.SetControllerReference(cr, pr, r.Scheme); err != nil {
		return &pov1.PrometheusRule{}, fmt.Errorf("unable to fill owner field for PrometheusRule '%v/%v': %v",
			pr.ObjectMeta.Namespace, pr.ObjectMeta.Name, err)
	}
	return pr, nil
}
