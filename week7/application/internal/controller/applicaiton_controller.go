/*
Copyright 2024.

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

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	applicationv1 "github.com/sawyer523/application/api/v1"
)

// ApplicaitonReconciler reconciles a Applicaiton object
type ApplicaitonReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=application.aiops.com,resources=applicaitons,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=application.aiops.com,resources=applicaitons/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=application.aiops.com,resources=applicaitons/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Applicaiton object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *ApplicaitonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var app applicationv1.Applicaiton
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		logger.Error(err, "unable to fetch Applicaiton")
		return ctrl.Result{}, nil
	}

	logger.Info("Reconcile Applicaiton", "Applicaiton", app.Name)

	labels := map[string]string{
		"app": app.Name,
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(
		ctx, r.Client, configMap, func() error {
			configMap.Data = app.Spec.ConfigMap.Data
			configMap.BinaryData = app.Spec.ConfigMap.BinaryData
			configMap.Immutable = app.Spec.ConfigMap.Immutable

			return controllerutil.SetOwnerReference(&app, configMap, r.Scheme)
		},
	)
	if err != nil {
		logger.Error(err, "unable to create or update ConfigMap")
		return ctrl.Result{}, err
	}

	logger.Info("ConfigMap "+string(op), "ConfigMap", configMap.Name)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}

	op, err = controllerutil.CreateOrUpdate(
		ctx, r.Client, deployment, func() error {
			replicas := int32(1)
			if app.Spec.Deployment.Replicas != 0 {
				replicas = app.Spec.Deployment.Replicas
			}
			deployment.Spec = appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  app.Name,
								Image: app.Spec.Deployment.Image,
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: app.Spec.Deployment.Port,
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      app.Name,
										MountPath: "/etc/config",
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: app.Name,
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: app.Name,
										},
									},
								},
							},
						},
					},
				},
			}

			return controllerutil.SetOwnerReference(&app, deployment, r.Scheme)
		},
	)

	if err != nil {
		logger.Error(err, "unable to create or update Deployment")
		return ctrl.Result{}, err
	}

	logger.Info("Deployment "+string(op), "Deployment", deployment.Name)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}

	op, err = controllerutil.CreateOrUpdate(
		ctx, r.Client, service, func() error {
			service.Spec = corev1.ServiceSpec{
				Selector: labels,
				Ports:    app.Spec.Service.Ports,
			}
			return controllerutil.SetOwnerReference(&app, service, r.Scheme)
		},
	)

	if err != nil {
		logger.Error(err, "unable to create or update Service")
		return ctrl.Result{}, err
	}

	logger.Info("Service "+string(op), "Service", service.Name)

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}

	op, err = controllerutil.CreateOrUpdate(
		ctx, r.Client, ingress, func() error {
			ingress.Spec = networkingv1.IngressSpec{
				IngressClassName: app.Spec.Ingress.IngressClassName,
				Rules:            app.Spec.Ingress.Rules,
			}
			return controllerutil.SetOwnerReference(&app, ingress, r.Scheme)
		},
	)

	if err != nil {
		logger.Error(err, "unable to create or update Ingress")
		return ctrl.Result{}, err
	}

	logger.Info("Ingress "+string(op), "Ingress", ingress.Name)

	// update status
	app.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	if err = r.Status().Update(ctx, &app); err != nil {
		logger.Error(err, "unable to update Applicaiton status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicaitonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&applicationv1.Applicaiton{}).
		Complete(r)
}
