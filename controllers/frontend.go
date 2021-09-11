package controllers

import (
	"context"
	appv1alpha1 "github.com/jxlwqq/hello-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const frontendImageName = "paulbouwer/hello-kubernetes"
const frontendDeployName = "hello"

func (r *HelloReconciler) frontendDeployment(h *appv1alpha1.Hello) *appsv1.Deployment {
	size := h.Spec.Size
	version := h.Spec.Version
	image := frontendImageName + ":" + version
	labels := labels("frontend")
	dep := &appsv1.Deployment{

		ObjectMeta: metav1.ObjectMeta{
			Namespace: h.Namespace,
			Name:      frontendDeployName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Name:            "hello",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "hello",
						}},
					}},
				},
			},
		},
	}

	_ = controllerutil.SetControllerReference(h, dep, r.Scheme)

	return dep
}

func (r *HelloReconciler) frontendService(h *appv1alpha1.Hello) *corev1.Service {
	labels := labels("frontend")
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: h.Namespace,
			Name:      "hello-svc",
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
				NodePort:   30691,
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	_ = controllerutil.SetControllerReference(h, svc, r.Scheme)

	return svc
}

func (r *HelloReconciler) handleFrontendChanges(h *appv1alpha1.Hello) (*ctrl.Result, error) {
	size := h.Spec.Size
	version := h.Spec.Version
	image := frontendImageName + ":" + version
	dep := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: h.Namespace,
		Name:      frontendDeployName,
	}, dep)
	if err != nil {
		return &ctrl.Result{}, err
	}
	if size != *dep.Spec.Replicas {
		*dep.Spec.Replicas = size
		err = r.Client.Update(context.TODO(), dep)
		if err != nil {
			return &ctrl.Result{}, err
		}
	}

	if image != (*dep).Spec.Template.Spec.Containers[0].Image {
		(*dep).Spec.Template.Spec.Containers[0].Image = image
		err = r.Client.Update(context.TODO(), dep)
		if err != nil {
			return &ctrl.Result{}, err
		}
	}

	return nil, nil
}
