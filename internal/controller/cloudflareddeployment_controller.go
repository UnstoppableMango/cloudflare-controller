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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cfv1alpha1 "github.com/UnstoppableMango/cloudflare-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CloudflaredDeploymentReconciler reconciles a CloudflaredDeployment object
type CloudflaredDeploymentReconciler struct {
	client.Client
	logger logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cloudflare.cloudflare.unmango.net,resources=cloudflareddeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloudflare.cloudflare.unmango.net,resources=cloudflareddeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloudflare.cloudflare.unmango.net,resources=cloudflareddeployments/finalizers,verbs=update

func (r *CloudflaredDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = log.FromContext(ctx)

	deployment := &cfv1alpha1.CloudflaredDeployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		r.logger.Error(err, "Failed to get CloudflaredDeployment")
		return ctrl.Result{}, err
	}

	if deployment.Spec.Kind == cfv1alpha1.DaemonSet {
		return reconcile[*appsv1.DaemonSet](r, ctx, req, deployment)
	}
	if deployment.Spec.Kind == cfv1alpha1.Deployment {
		return r.reconcileDeployment(ctx, req, deployment)
	}

	r.logger.Info("Invalid CloudflaredDeployment kind")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudflaredDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cfv1alpha1.CloudflaredDeployment{}).
		Complete(r)
}

const (
	defaultImage string = "docker.io/cloudflare/cloudflared:latest"
	defaultName  string = "cloudflared"
)

var defaultMatchLabels = map[string]string{"app": "cloudflared"}
var defaultSelector = metav1.LabelSelector{MatchLabels: defaultMatchLabels}
var defaultTemplate = v1.PodTemplateSpec{
	ObjectMeta: metav1.ObjectMeta{Labels: defaultMatchLabels},
	Spec: v1.PodSpec{
		Containers: []v1.Container{{
			Name:  defaultName,
			Image: defaultImage,
		}},
	},
}

func defaultDaemonSet(deployment cfv1alpha1.CloudflaredDeployment) appsv1.DaemonSet {
	return appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: deployment.Namespace,
			Name:      deployment.Name,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &defaultSelector,
			Template: defaultTemplate,
		},
	}
}

func defaultDeployment(deployment cfv1alpha1.CloudflaredDeployment) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: deployment.Namespace,
			Name:      deployment.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &defaultSelector,
			Template: defaultTemplate,
		},
	}
}

type kind interface {
	runtime.Object
	*appsv1.DaemonSet | *appsv1.Deployment
}

type defaulter[T kind] interface {
	Default(deployment *cfv1alpha1.CloudflaredDeployment) T
}

func reconcile[T kind](
	r *CloudflaredDeploymentReconciler,
	ctx context.Context,
	req ctrl.Request,
	deployment *cfv1alpha1.CloudflaredDeployment,
	defaulter defaulter[T],
) (ctrl.Result, error) {
	var app T
	err := r.Get(ctx, req.NamespacedName, app)
	if err == nil {
		r.logger.Info("DaemonSet is up to date")
		return ctrl.Result{}, nil
	}

	if !errors.IsNotFound(err) {
		r.logger.Error(err, "Failed to get DaemonSet")
		return ctrl.Result{}, err
	}

	app = defaulter.Default(deployment)
	if deployment.Spec.Template != nil {
		deployment.Spec.Template.DeepCopyInto(&app.Spec.Template)
	}

	err = r.Create(ctx, app)
	if err != nil {
		r.logger.Error(err, "Failed to create DaemonSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CloudflaredDeploymentReconciler) reconcileDaemonSet(
	ctx context.Context,
	req ctrl.Request,
	deployment *cfv1alpha1.CloudflaredDeployment,
) (ctrl.Result, error) {
	app := &appsv1.DaemonSet{}
	err := r.Get(ctx, req.NamespacedName, app)
	if err == nil {
		r.logger.Info("DaemonSet is up to date")
		return ctrl.Result{}, nil
	}

	if !errors.IsNotFound(err) {
		r.logger.Error(err, "Failed to get DaemonSet")
		return ctrl.Result{}, err
	}

	app.Namespace = deployment.Namespace
	app.Name = deployment.Name
	app.Spec = appsv1.DaemonSetSpec{
		Selector: &defaultSelector,
		Template: defaultTemplate,
	}

	if deployment.Spec.Template != nil {
		deployment.Spec.Template.DeepCopyInto(&app.Spec.Template)
	}

	err = r.Create(ctx, app)
	if err != nil {
		r.logger.Error(err, "Failed to create DaemonSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CloudflaredDeploymentReconciler) reconcileDeployment(
	ctx context.Context,
	req ctrl.Request,
	deployment *cfv1alpha1.CloudflaredDeployment,
) (ctrl.Result, error) {
	app := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, app)
	if err == nil {
		r.logger.Info("Deployment is up to date")
		return ctrl.Result{}, nil
	}

	if !errors.IsNotFound(err) {
		r.logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	app.Namespace = deployment.Namespace
	app.Name = deployment.Name
	matchLabels := map[string]string{"app": "cloudflared"}
	app.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{MatchLabels: matchLabels},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: matchLabels},
			Spec: v1.PodSpec{
				Containers: []v1.Container{{
					Name:  "cloudflared",
					Image: defaultImage,
				}},
			},
		},
	}

	if deployment.Spec.Template != nil {
		deployment.Spec.Template.DeepCopyInto(&app.Spec.Template)
	}

	err = r.Create(ctx, app)
	if err != nil {
		r.logger.Error(err, "Failed to create DaemonSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
