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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1alpha1 "github.com/jxlwqq/hello-operator/api/v1alpha1"
)

// HelloReconciler reconciles a Hello object
type HelloReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.jxlwqq.github.io,resources=hellos,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.jxlwqq.github.io,resources=hellos/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.jxlwqq.github.io,resources=hellos/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Hello object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *HelloReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)
	reqLogger.Info("Reconciling Hello")

	hello := &appv1alpha1.Hello{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, hello)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var result *ctrl.Result
	reqLogger.Info("---Frontend Deployment---")
	result, err = r.ensureDeployment(r.frontendDeployment(hello))
	if result != nil {
		return *result, err
	}

	reqLogger.Info("---Frontend Service---")
	result, err = r.ensureService(r.frontendService(hello))
	if result != nil {
		return *result, err
	}

	reqLogger.Info("---Frontend Change Handler---")
	result, err = r.handleFrontendChanges(hello)
	if result != nil {
		return *result, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Hello{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
