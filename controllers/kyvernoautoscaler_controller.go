/*
Copyright 2023.

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

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	autoscalerv1alpha1 "kyverno-autoscaler/api/v1alpha1"
)

// KyvernoAutoscalerReconciler reconciles a KyvernoAutoscaler object
type KyvernoAutoscalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=autoscaler.security.nirmata.io,resources=kyvernoautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaler.security.nirmata.io,resources=kyvernoautoscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=autoscaler.security.nirmata.io,resources=kyvernoautoscalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KyvernoAutoscaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *KyvernoAutoscalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	logger := log.FromContext(ctx)
	logger.Info("reconciling event...")

	autoscaler := &autoscalerv1alpha1.KyvernoAutoscaler{}
	err := r.Get(ctx, req.NamespacedName, autoscaler)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
	}

	desiredReplicas := computeDesiredReplicas(autoscaler)

	deploy := &v1.Deployment{}
	helper, _ := patch.NewHelper(deploy, r.Client)
	defer func() {
		if err = helper.Patch(ctx, deploy); err != nil && reterr == nil {
			logger.Error(err, "failed to patch kyverno deplpyment")
			reterr = err
		}
	}()

	kyvernoDeployment := types.NamespacedName{
		Namespace: "kyverno",
		Name:      "kyverno",
	}
	err = r.Get(ctx, kyvernoDeployment, deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
	}

	if *deploy.Spec.Replicas != desiredReplicas {
		*deploy.Spec.Replicas = desiredReplicas
		helper.Patch(ctx, deploy)
	}

	return ctrl.Result{}, nil
}

func computeDesiredReplicas(autoscaler *autoscalerv1alpha1.KyvernoAutoscaler) int32 {

	// write more complex logic
	// if ARPS > 100, scale kyverno to 3 replicas
	if autoscaler.Spec.Arps > 100 {
		return 3
	}

	return 1
}

// SetupWithManager sets up the controller with the Manager.
func (r *KyvernoAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalerv1alpha1.KyvernoAutoscaler{}).
		Complete(r)
}
