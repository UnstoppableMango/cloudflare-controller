/*
Copyright 2024 UnstoppableMango.

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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cfv1alpha1 "github.com/UnstoppableMango/cloudflare-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	defaultImage string = "docker.io/cloudflare/cloudflared:latest"
)

// CloudflaredDeploymentReconciler reconciles a CloudflaredDeployment object
type CloudflaredDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cloudflare.cloudflare.unmango.net,resources=cloudflareddeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloudflare.cloudflare.unmango.net,resources=cloudflareddeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloudflare.cloudflare.unmango.net,resources=cloudflareddeployments/finalizers,verbs=update

func (r *CloudflaredDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	deployment := &cfv1alpha1.CloudflaredDeployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		logger.Error(err, "Failed to get CloudflaredDeployment")
		return ctrl.Result{}, err
	}

	if deployment.Spec.Kind == cfv1alpha1.DaemonSet {
		err = r.Get(ctx, req.NamespacedName, &appsv1.DaemonSet{})
	} else if deployment.Spec.Kind == cfv1alpha1.Deployment {
		err = r.Get(ctx, req.NamespacedName, &appsv1.Deployment{})
	} else {
		logger.Info("Invalid CloudflaredDeployment kind")
		return ctrl.Result{}, nil
	}
	if err == nil {
		logger.Info("Up to date", "kind", deployment.Spec.Kind)
		return ctrl.Result{}, nil
	}

	if !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get", "kind", deployment.Spec.Kind)
		return ctrl.Result{}, err
	}

	var app client.Object
	switch deployment.Spec.Kind {
	case cfv1alpha1.DaemonSet:
		app = deployment.ToDaemonSet(defaultImage)
	case cfv1alpha1.Deployment:
		app = deployment.ToDeployment(defaultImage)
	}

	err = r.Create(ctx, app)
	if err != nil {
		logger.Error(err, "Failed to create", "kind", deployment.Spec.Kind)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflaredDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cfv1alpha1.CloudflaredDeployment{}).
		Complete(r)
}
