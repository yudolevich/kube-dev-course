/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
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
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	deployv1alpha1 "github.com/yudolevich/kube-dev-course/example/operator/api/v1alpha1"
)

// NginxReconciler reconciles a Nginx object
type NginxReconciler struct {
	client.Client
	Tbot   *TBot
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=deploy.miit.ru,resources=nginxes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=deploy.miit.ru,resources=nginxes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=deploy.miit.ru,resources=nginxes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Nginx object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	nginx := &deployv1alpha1.Nginx{}
	if err := r.Get(ctx, req.NamespacedName, nginx); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	logger.WithValues("name", nginx.GetName(), "namespace", nginx.GetNamespace())

	logger.Info("reconicle")
	if !nginx.Status.Approved {
		r.Tbot.SendDeploy(nginx)
		return ctrl.Result{}, nil
	}

	if err := r.deploy(ctx, nginx); err != nil {
		logger.Error(err, "error deploy nginx")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&deployv1alpha1.Nginx{}).
		Complete(r)
}

func (r *NginxReconciler) deploy(ctx context.Context, nginx *deployv1alpha1.Nginx) error {
	labels := map[string]string{"nginx": nginx.GetName()}
	owner := metav1.OwnerReference{
		APIVersion: nginx.APIVersion,
		Kind:       nginx.Kind,
		UID:        nginx.GetUID(),
		Name:       nginx.GetName(),
	}

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            nginx.GetName(),
			Namespace:       nginx.GetNamespace(),
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{owner},
		},
		Data: map[string]string{
			"index.html": nginx.Spec.Index,
		},
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            nginx.GetName(),
			Namespace:       nginx.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{owner},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Replicas: &nginx.Spec.Replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "index",
									MountPath: "/usr/share/nginx/html",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "index",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: nginx.GetName(),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := r.apply(ctx, cm); err != nil {
		return err
	}

	if err := r.apply(ctx, deploy); err != nil {
		return err
	}

	return nil
}

func (r *NginxReconciler) apply(ctx context.Context, obj client.Object) error {
	logger := log.FromContext(ctx)

	if err := r.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		logger.Info("create")
		if err := r.Create(ctx, obj); err != nil {
			return err
		}

		return nil
	}

	logger.Info("update")
	if err := r.Update(ctx, obj); err != nil {
		return err
	}

	return nil
}
