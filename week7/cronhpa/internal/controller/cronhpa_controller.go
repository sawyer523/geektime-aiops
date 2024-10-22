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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/robfig/cron/v3"
	autoscalingv1 "github.com/sawyer523/cronhpa/api/v1"
)

// CronHPAReconciler reconciles a CronHPA object
type CronHPAReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=autoscaling.aiops.com,resources=cronhpas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.aiops.com,resources=cronhpas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=autoscaling.aiops.com,resources=cronhpas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CronHPA object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *CronHPAReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling CronHPA")
	var cornHPA autoscalingv1.CronHPA
	if err := r.Get(ctx, req.NamespacedName, &cornHPA); err != nil {
		logger.Error(err, "unable to fetch CronHPA")
		return ctrl.Result{}, nil
	}

	if cornHPA.Spec.ConfigMap != nil {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cornHPA.Name,
				Namespace: cornHPA.Namespace,
			},
			Data:       cornHPA.Spec.ConfigMap.Data,
			BinaryData: cornHPA.Spec.ConfigMap.BinaryData,
			Immutable:  cornHPA.Spec.ConfigMap.Immutable,
		}

		// Create or update the CronHPA ConfigMap
		res, err := controllerutil.CreateOrUpdate(
			ctx, r.Client, cm, func() error {
				return controllerutil.SetOwnerReference(&cornHPA, cm, r.Scheme)
			},
		)
		if err != nil {
			logger.Error(err, "failed to create or update ConfigMap")
			return ctrl.Result{}, nil
		}

		if res != controllerutil.OperationResultNone {
			logger.Info("ConfigMap " + string(res) + " for CronHPA")
		}

	}

	now := time.Now()
	var earliestNextRunTime *time.Time
	for _, job := range cornHPA.Spec.Jobs {
		lastRunTime := cornHPA.Status.LastRuntime[job.Name]
		nextScheduledTime, err := r.getNextScheduleTime(job.Schedule, lastRunTime.Time)
		if err != nil {
			logger.Error(err, "failed to get next schedule time")
			return ctrl.Result{}, err
		}

		logger.Info("Job info", "job", job.Name, "lastRunTime", lastRunTime, "nextScheduledTime", nextScheduledTime)

		if now.After(nextScheduledTime) || now.Equal(nextScheduledTime) {
			logger.Info("Update replicas", "job", job.Name, "targetSize", job.TargetSize)
			if err = r.updateDeploymentReplicas(ctx, &cornHPA, cornHPA.Spec.ScaleTargetRef, job); err != nil {
				logger.Error(err, "failed to update deployment replicas")
				return ctrl.Result{}, err
			}

			cornHPA.Status.CurrentReplicas = job.TargetSize
			cornHPA.Status.LastScaleTime = &metav1.Time{Time: now}

			if cornHPA.Status.LastRuntime == nil {
				cornHPA.Status.LastRuntime = make(map[string]metav1.Time)
			}
			cornHPA.Status.LastRuntime[job.Name] = metav1.Time{Time: now}

			nextRunTime, _ := r.getNextScheduleTime(job.Schedule, now)
			if earliestNextRunTime == nil || nextRunTime.Before(*earliestNextRunTime) {
				earliestNextRunTime = &nextRunTime
			}
		} else {
			if earliestNextRunTime == nil || nextScheduledTime.Before(*earliestNextRunTime) {
				earliestNextRunTime = &nextScheduledTime
			}
		}
	}

	if err := r.Status().Update(ctx, &cornHPA); err != nil {
		logger.Error(err, "failed to update CronHPA status")
		return ctrl.Result{}, err
	}

	if earliestNextRunTime != nil {
		requeueAfter := earliestNextRunTime.Sub(now)
		if requeueAfter < 0 {
			requeueAfter = time.Second
		}
		logger.Info("Requeue after", "duration", requeueAfter)
		return reconcile.Result{RequeueAfter: requeueAfter}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronHPAReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalingv1.CronHPA{}).
		Complete(r)
}

func (r *CronHPAReconciler) getNextScheduleTime(schedule string, after time.Time) (time.Time, error) {
	parse := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	cronSchedule, err := parse.Parse(schedule)
	if err != nil {
		return time.Time{}, err
	}
	return cronSchedule.Next(after), nil
}

func (r *CronHPAReconciler) updateDeploymentReplicas(
	ctx context.Context,
	cronHPA *autoscalingv1.CronHPA,
	ref autoscalingv1.ScaleTargetRefrence,
	job autoscalingv1.JobSpec,
) error {
	logger := log.FromContext(ctx)

	deployment := appsv1.Deployment{}
	deploymentKey := types.NamespacedName{
		Namespace: cronHPA.Namespace,
		Name:      cronHPA.Name,
	}

	if err := r.Get(ctx, deploymentKey, &deployment); err != nil {
		logger.Error(err, "failed to get deployment")
		return err
	}

	if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == job.TargetSize {
		logger.Info("Deployment replicas already set to target size", "targetSize", job.TargetSize)
		return nil
	}

	deployment.Spec.Replicas = &job.TargetSize
	if err := r.Update(ctx, &deployment); err != nil {
		logger.Error(err, "failed to update deployment")
		return err
	}

	logger.Info("Deployment replicas updated", "targetSize", job.TargetSize)
	return nil
}
