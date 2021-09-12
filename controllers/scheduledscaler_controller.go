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
	"github.com/go-logr/logr"
	scalerv1alpha1 "github.com/kmdrn7/scheduled-scaler-operator/api/v1alpha1"
	timeUtils "github.com/kmdrn7/scheduled-scaler-operator/pkg/time"
	appsv1 "k8s.io/api/apps/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	//v1 "k8s.io/client-go/applyconfigurations/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ScheduledScalerReconciler reconciles a ScheduledScaler object
type ScheduledScalerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=scaler.andikahmadr.io,resources=scheduledscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scaler.andikahmadr.io,resources=scheduledscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=scaler.andikahmadr.io,resources=scheduledscalers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ScheduledScaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *ScheduledScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log
	cl, errCl := client.New(config.GetConfigOrDie(), client.Options{})
	if errCl != nil {
		log.Error(errCl, "Failed to create client")
		os.Exit(1)
	}

	instance := &scalerv1alpha1.ScheduledScaler{}
	errInstance := r.Get(ctx, req.NamespacedName, instance)
	if errInstance != nil {
		if errors.IsNotFound(errInstance) {
			log.Info("ScheduledScaler resource not found.")
			return ctrl.Result{}, nil
		}
		log.Error(errInstance, "Error getting resource.")
		return ctrl.Result{}, errInstance
	}
	log.Info("Reconciling resource " + req.NamespacedName.String())

	if instance.Status.StoredReplicaCount == 0 {
		instance.Status.StoredReplicaCount = -1
	}

	if instance.Status.Phase == "" {
		log.Info("Resource's phase is not initiated, changed to pending phase")
		instance.Status.Phase = scalerv1alpha1.PhasePending
	} else if instance.Status.Phase == scalerv1alpha1.PhaseDone {
		instance.Status.Phase = scalerv1alpha1.PhaseRunning
	}

	now := timeUtils.Now()
	scheduleStart := instance.Spec.Schedule.Start
	scheduleEnd := instance.Spec.Schedule.End

	timeStart, err := timeUtils.Parse(scheduleStart)
	if err != nil {
		fmt.Println("Error parsing start time")
	}

	timeEnd, err := timeUtils.Parse(scheduleEnd)
	if err != nil {
		fmt.Println("Error parsing end time")
	}

	// state transition PENDING -> RUNNING -> DONE
	switch instance.Status.Phase {
	case scalerv1alpha1.PhasePending:
		log.Info(req.NamespacedName.String() + " still in pending phase")
		log.Info(req.NamespacedName.String()+" detail", "now", now.String(), "time start", timeStart.String(), "stored replica count", strconv.Itoa(int(instance.Status.StoredReplicaCount)))

		if now.Before(timeStart) {
			reconcileAfter := timeStart.Sub(now)
			log.Info("Time to reconcile is " + reconcileAfter.String())
			return ctrl.Result{RequeueAfter: reconcileAfter}, nil
		}

		log.Info("change " + req.NamespacedName.String() + " phase to running")
		instance.Status.Phase = scalerv1alpha1.PhaseRunning

	case scalerv1alpha1.PhaseRunning:
		log.Info(req.NamespacedName.String() + " is in running phase")
		log.Info(req.NamespacedName.String()+" detail", "now", now.String(), "time start", timeStart.String(), "stored replica count", strconv.Itoa(int(instance.Status.StoredReplicaCount)))

		// get deployment
		deployment := &appsv1.Deployment{}
		err := cl.Get(ctx, client.ObjectKey{
			Namespace: instance.Namespace,
			Name:      instance.Spec.DeploymentName,
		}, deployment)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Error(err, "Cannot find deployment with name "+deployment.Name+"in "+instance.Namespace+" namespace")
				log.Error(err, "Trying again after 10 seconds")
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}
		}

		if instance.Status.StoredReplicaCount == -1 {
			instance.Status.StoredReplicaCount = *deployment.Spec.Replicas // store, and reconcile
			break
		}

		if now.After(timeStart) && now.Before(timeEnd) {
			// get deployment's replica count and compare with instance's spec
			// update status StoredReplicaCount to match with Deployment's replica count
			log.Info("Deployment " + deployment.Name + " has " + strconv.Itoa(int(*deployment.Spec.Replicas)) + " replicas")
			if *deployment.Spec.Replicas != instance.Spec.ReplicaCount {
				log.Info("Replica count in deployment " + deployment.Name + " didn't match with " + req.NamespacedName.String())
				log.Info("Scaling deployment " + deployment.Name + " to " + strconv.Itoa(int(instance.Spec.ReplicaCount)) + " replicas")
				deployment.Spec.Replicas = &instance.Spec.ReplicaCount
				err := cl.Update(ctx, deployment)
				if err != nil {
					log.Error(err, "Error updating deployment "+deployment.Name)
					return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
				}
				log.Info("Successfully scaling " + deployment.Name + " with " + strconv.Itoa(int(instance.Spec.ReplicaCount)) + " replicas")
			}
			// reconcile
			reconcileAfter := timeEnd.Sub(now)
			log.Info("Time to reconcile is " + reconcileAfter.String())
			return ctrl.Result{RequeueAfter: reconcileAfter}, nil
		} else {
			log.Info("change phase to done")
			instance.Status.Phase = scalerv1alpha1.PhaseDone
		}

	case scalerv1alpha1.PhaseDone:
		log.Info(req.NamespacedName.String() + " is in done phase")
		log.Info(req.NamespacedName.String()+" detail", "now", now.String(), "time start", timeStart.String(), "stored replica count", strconv.Itoa(int(instance.Status.StoredReplicaCount)))

		// get deployment
		deployment := &appsv1.Deployment{}
		err := cl.Get(ctx, client.ObjectKey{
			Namespace: instance.Namespace,
			Name:      instance.Spec.DeploymentName,
		}, deployment)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Error(err, "Cannot find deployment with name "+deployment.Name+"in "+instance.Namespace+" namespace")
				log.Error(err, "Trying again after 10 seconds")
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}
		}
		log.Info("Deployment " + deployment.Name + " has " + strconv.Itoa(int(*deployment.Spec.Replicas)) + " replicas")

		if *deployment.Spec.Replicas == instance.Status.StoredReplicaCount {
			log.Info("Nothing to do since deployment's replicas is not changed")
		} else {
			deployment.Spec.Replicas = &instance.Status.StoredReplicaCount
			err = cl.Update(ctx, deployment)
			if err != nil {
				log.Error(err, "Error updating deployment "+deployment.Name)
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}
			log.Info("Successfully scaling back " + deployment.Name + " to " + strconv.Itoa(int(instance.Status.StoredReplicaCount)) + " replicas")
		}

	default:
		return ctrl.Result{}, nil
	}

	// update status
	errLast := r.Status().Update(ctx, instance)
	if errLast != nil {
		return ctrl.Result{}, errLast
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScheduledScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scalerv1alpha1.ScheduledScaler{}).
		Complete(r)
}
