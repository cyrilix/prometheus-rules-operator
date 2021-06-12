package controllers

import (
	"context"
	"github.com/cyrilix/prometheus-rules-operator/api/v1alpha1"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:docs-gen:collapse=Imports

// Define utility constants for object names and testing timeouts/durations and intervals.
const (
	CRName = "test-rules"

	timeout  = time.Second * 5
	interval = time.Millisecond * 250
)

var _ = Describe("Rules controller", func() {

	Context("When create CR with default values", func() {
		namespace := "default-cr"
		var createdRules v1alpha1.Rule
		var createdPR pov1.PrometheusRule

		It("Should create new namespace without errors", func() {
			ctx := context.Background()
			Expect(k8sClient.Create(ctx, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			})).Should(Succeed())
		})

		It("Should create Custom resources without errors", func() {
			ctx := context.Background()
			rule := &v1alpha1.Rule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      CRName,
					Namespace: namespace,
				},
				Spec: v1alpha1.RuleSpec{},
			}
			Expect(k8sClient.Create(ctx, rule)).Should(Succeed())

			ruleLookupKey := types.NamespacedName{Namespace: namespace, Name: rule.Name}

			// We'll need to retry getting this newly created resource, given that creation may not immediately happen.
			Eventually(func() error {
				return k8sClient.Get(ctx, ruleLookupKey, &createdRules)
			}, timeout, interval).Should(Succeed())
		})

		It("Should create new PrometheusRules", func() {

			ctx := context.Background()
			prLookupKey := types.NamespacedName{Namespace: namespace, Name: CRName}

			By("Creating a new Deployment", func() {
				Eventually(func() error {
					return k8sClient.Get(ctx, prLookupKey, &createdPR)
				}, timeout, interval).Should(Succeed())
			})

			By("Checking owner field is set", func() {
				Expect(createdPR.OwnerReferences[0].UID).To(Equal(createdRules.UID))
			})

			By("Checking many rules exist", func() {
				Expect(createdPR.Spec.Groups).ToNot(BeEmpty())
			})
		})
	})
})
