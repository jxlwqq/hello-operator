package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *HelloReconciler) ensureDeployment(dep *appsv1.Deployment) (*ctrl.Result, error) {
	found := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: dep.Namespace,
		Name:      dep.Name,
	}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), dep)
			if err != nil {
				return &ctrl.Result{}, err
			}
		}
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func (r *HelloReconciler) ensureService(svc *corev1.Service) (*ctrl.Result, error) {
	found := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: svc.Namespace,
		Name:      svc.Name,
	}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), svc)
			if err != nil {
				return &ctrl.Result{}, err
			}
		}
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func labels(tier string) map[string]string {
	return map[string]string{
		"app":  "hello",
		"tier": tier,
	}
}
