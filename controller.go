/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"os"
	"context"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	samplev1alpha1 "k8s.io/sample-controller/pkg/apis/samplecontroller/v1alpha1"
	samplescheme "k8s.io/sample-controller/pkg/client/clientset/versioned/scheme"
)

func main() {
	mgr, err := builder.SimpleController().
		ForType(&samplev1alpha1.Foo{}).
		Owns(&appsv1.Deployment{}).
		Build(&Controller{})

	if err != nil {
		logf.Log.Error(err, "unable to construct foo controller")
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logf.Log.Error(err, "unable to start controllers")
		os.Exit(1)
	}
}

type Controller struct {
	client.Client
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log := logf.Log.WithName("foo-controller").WithValues("foo", req.NamespacedName)
	// Get the Foo resource with this namespace/name
	var foo samplev1alpha1.Foo
	if err := c.Get(context.Background(), req.NamespacedName, &foo); err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, "foo no longer exists")
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	deploymentName := foo.Spec.DeploymentName
	if deploymentName == "" {
		log.Error(nil, "deployment name must be specified")
		return reconcile.Result{}, nil
	}

	var deployment appsv1.Deployment
	_, err := controllerutil.CreateOrUpdate(context.Background(), c, &deployment, func(existing runtime.Object) error {
		depl := existing.(*appsv1.Deployment)
		depl.Name = deploymentName
		depl.Namespace = foo.Namespace
		if err := controllerutil.SetControllerReference(&foo, depl, samplescheme.Scheme); err != nil {
			return err
		}
		labels := map[string]string{
			"app":        "nginx",
			"controller": foo.Name,
		}
		depl.Spec = appsv1.DeploymentSpec{
			Replicas: foo.Spec.Replicas,
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
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
			},
		}
		return nil
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	foo.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	if err := c.Update(context.Background(), &foo); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (c *Controller) InjectClient(cl client.Client) error {
	c.Client = cl
	return nil
}
