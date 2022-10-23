package main

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mygroupv1alpha1 "github.com/myid/myresource-crd/pkg/apis/mygroup.example.com/v1alpha1"
)

const (
	Name = "MyResourceReconciler"

	_buildingState = "Building"
	_readyState    = "Ready"
)

type MyReconciler struct {
	client        client.Client
	EventRecorder record.EventRecorder
}

func (a *MyReconciler) InjectClient(
	c client.Client,
) error {
	a.client = c
	return nil
}

func (a *MyReconciler) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	log := log.FromContext(ctx)

	log.Info("getting myresource instance")

	myresource := mygroupv1alpha1.MyResource{}
	err := a.client.Get( // ❶
		ctx,
		req.NamespacedName,
		&myresource,
		&client.GetOptions{},
	)
	if err != nil {
		if errors.IsNotFound(err) { // ❷
			log.Info("resource is not found")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	ownerReference := metav1.NewControllerRef( // ❸
		&myresource,
		mygroupv1alpha1.SchemeGroupVersion.
			WithKind("MyResource"),
	)

	err = a.applyDeployment( // ❹
		ctx,
		&myresource,
		ownerReference,
	)
	if err != nil {
		return reconcile.Result{}, err
	}

	status, err := a.computeStatus(ctx, &myresource) // ❺
	if err != nil {
		return reconcile.Result{}, err
	}
	myresource.Status = *status
	log.Info("updating status", "state", status.State)
	err = a.client.Status().Update(ctx, &myresource) // ❻
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (a *MyReconciler) applyDeployment(
	ctx context.Context,
	myres *mygroupv1alpha1.MyResource,
	ownerref *metav1.OwnerReference,
) error {
	deploy := createDeployment(myres, ownerref)
	err := a.client.Patch( // ❼
		ctx,
		deploy,
		client.Apply,
		client.FieldOwner(Name),
		client.ForceOwnership,
	)
	return err
}

func createDeployment(
	myres *mygroupv1alpha1.MyResource,
	ownerref *metav1.OwnerReference,
) *appsv1.Deployment {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"myresource": myres.GetName(),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"myresource": myres.GetName(),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"myresource": myres.GetName(),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: myres.Spec.Image, // ❽
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: myres.Spec.Memory, // ❾
								},
							},
						},
					},
				},
			},
		},
	}
	deploy.SetName(myres.GetName() + "-deployment")
	deploy.SetNamespace(myres.GetNamespace())
	deploy.SetGroupVersionKind(
		appsv1.SchemeGroupVersion.WithKind("Deployment"),
	)
	deploy.SetOwnerReferences([]metav1.OwnerReference{ // ❿
		*ownerref,
	})
	return deploy
}

func (a *MyReconciler) computeStatus(
	ctx context.Context,
	myres *mygroupv1alpha1.MyResource,
) (*mygroupv1alpha1.MyResourceStatus, error) {

	logger := log.FromContext(ctx)
	result := mygroupv1alpha1.MyResourceStatus{
		State: _buildingState,
	}

	deployList := appsv1.DeploymentList{}
	err := a.client.List( // ⓫
		ctx,
		&deployList,
		client.InNamespace(myres.GetNamespace()),
		client.MatchingLabels{
			"myresource": myres.GetName(),
		},
	)
	if err != nil {
		return nil, err
	}

	if len(deployList.Items) == 0 {
		logger.Info("no deployment found")
		return &result, nil
	}

	if len(deployList.Items) > 1 {
		logger.Info(
			"too many deployments found", "count",
			len(deployList.Items),
		)
		return nil, fmt.Errorf(
			"%d deployment found, expected 1",
			len(deployList.Items),
		)
	}

	status := deployList.Items[0].Status // ⓬
	logger.Info(
		"got deployment status",
		"status", status,
	)
	if status.ReadyReplicas == 1 {
		result.State = _readyState // ⓭
	}

	return &result, nil
}
