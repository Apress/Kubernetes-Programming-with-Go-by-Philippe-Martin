/*
Copyright 2022.

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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mygroupv1alpha1 "github.com/myid/myresource/api/v1alpha1"
)

const (
	Name = "myresource-controller"
)

// MyResourceReconciler reconciles a MyResource object
type MyResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mygroup.myid.dev,resources=myresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mygroup.myid.dev,resources=myresources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mygroup.myid.dev,resources=myresources/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyResource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *MyResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("getting myresource instance")

	myresource := mygroupv1alpha1.MyResource{}
	err := r.Client.Get(
		ctx,
		req.NamespacedName,
		&myresource,
		&client.GetOptions{},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("resource is not found")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	ownerReference := metav1.NewControllerRef(
		&myresource,
		mygroupv1alpha1.GroupVersion.WithKind("MyResource"),
	)

	err = r.applyDeployment(ctx, &myresource, ownerReference)
	if err != nil {
		return reconcile.Result{}, err
	}

	status, err := r.computeStatus(ctx, &myresource)
	if err != nil {
		return reconcile.Result{}, err
	}
	myresource.Status = *status
	log.Info("updating status", "state", status.State)
	err = r.Client.Status().Update(ctx, &myresource)
	if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mygroupv1alpha1.MyResource{}).
		Complete(r)
}
