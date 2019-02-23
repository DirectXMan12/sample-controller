/*
Copyright 2019 The Kubernetes Authors.

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

//go:generate go run ./vendor/sigs.k8s.io/controller-tools/cmd/crd generate --domain metamagical.io
package main

import (
	"os"
	"context"
	"time"
	"math/rand"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	 _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	api "k8s.io/sample-controller/pkg/apis/chaosapps/v1"
)

var (
	setupLog = ctrl.Log.WithName("setup")
	recLog = ctrl.Log.WithName("reconciler")
)

type reconciler struct {
	client.Client

	scheme *runtime.Scheme
}

func (r *reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := recLog.WithValues("chaospod", req.NamespacedName)
	log.V(1).Info("reconciling chaos pod")
	ctx := context.Background()

	var chaosctl api.ChaosPod
	if err := r.Get(ctx, req.NamespacedName, &chaosctl); err != nil {
		log.Error(err, "unable to get chaosctl")
		return ctrl.Result{}, err
	}

	var pod corev1.Pod
	podFound := true
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "unable to get pod")
			return ctrl.Result{}, err }
		podFound = false
	}

	if podFound {
		shouldStop := chaosctl.Spec.NextStop.Time.Before(time.Now())
		if !shouldStop {
			return ctrl.Result{RequeueAfter: chaosctl.Spec.NextStop.Sub(time.Now())+1*time.Second}, nil
		}

		if err := r.Delete(ctx, &pod); err != nil {
			log.Error(err, "unable to delete pod")
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	templ := chaosctl.Spec.Template.DeepCopy()
	pod.ObjectMeta = templ.ObjectMeta
	pod.Name = req.Name
	pod.Namespace = req.Namespace
	pod.Spec = templ.Spec

	if err := ctrl.SetControllerReference(&chaosctl, &pod, r.scheme); err != nil {
		log.Error(err, "unable to set pod's owner reference")
		return ctrl.Result{}, err
	}

	if err := r.Create(ctx, &pod); err != nil {
		log.Error(err, "unable to create pod")
		return ctrl.Result{}, err
	}

	chaosctl.Spec.NextStop.Time = time.Now().Add(time.Duration(10*(rand.Int63n(2)+1))*time.Second)
	chaosctl.Status.LastRun = pod.CreationTimestamp
	if err := r.Update(ctx, &chaosctl); err != nil {
		log.Error(err, "unable to update chaosctl status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func main() {
	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// in a real controller, we'd create a new scheme for this
	api.AddToScheme(mgr.GetScheme())

	err = ctrl.NewControllerManagedBy(mgr).
		For(&api.ChaosPod{}).
		Owns(&corev1.Pod{}).
		Complete(&reconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})

	if err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
