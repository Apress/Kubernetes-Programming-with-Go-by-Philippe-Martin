package main

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mygroupv1alpha1 "github.com/myid/myresource-crd/pkg/apis/mygroup.example.com/v1alpha1"
)

func main() {

	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	mygroupv1alpha1.AddToScheme(scheme)

	mgr, err := manager.New(
		config.GetConfigOrDie(),
		manager.Options{
			Scheme: scheme,
		},
	)
	panicIf(err)

	err = builder.
		ControllerManagedBy(mgr).
		For(&mygroupv1alpha1.MyResource{}).
		Owns(&corev1.Pod{}).
		Complete(&MyReconciler{})
	panicIf(err)

	err = mgr.Start(context.Background())
	panicIf(err)
}

type MyReconciler struct {
	client client.Client
}

func (a *MyReconciler) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	fmt.Printf("reconcile %v\n", req)

	// ## Getting information about a resource
	myresource := mygroupv1alpha1.MyResource{}
	err := a.client.Get(
		ctx,
		req.NamespacedName,
		&myresource,
	)
	if err != nil {
		fmt.Printf("%v\n", err)
		return reconcile.Result{}, err
	}

	// ## Listing resources
	resourcesList := mygroupv1alpha1.MyResourceList{}
	err = a.client.List(ctx, &resourcesList, client.InNamespace(req.Namespace))
	if err != nil {
		fmt.Printf("%v\n", err)
		return reconcile.Result{}, err
	}
	for _, res := range resourcesList.Items {
		fmt.Printf("res: %s\n", res.GetName())
	}

	// ## Creating a resource
	podToCreate := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "main",
					Image: "nginx",
				},
			},
		},
	}
	podToCreate.SetName(req.Name)
	podToCreate.SetNamespace(req.Namespace)
	err = a.client.Create(ctx, &podToCreate)
	if err != nil {
		fmt.Printf("create: %v\n", err)
		return reconcile.Result{}, err
	}

	// ## Deleting a resource
	podToDelete := corev1.Pod{}
	podToDelete.SetName(req.Name)
	podToDelete.SetNamespace(req.Namespace)
	err = a.client.Delete(ctx, &podToDelete)
	if err != nil {
		fmt.Printf("delete: %v\n", err)
		return reconcile.Result{}, err
	}

	// ## Deleting a collection of resources
	err = a.client.DeleteAllOf(
		ctx,
		&corev1.Pod{},
		client.InNamespace(req.Namespace))
	if err != nil {
		fmt.Printf("deleteAllOf: %v\n", err)
		// return reconcile.Result{}, err
	}

	// ## Patching a resource
	// ### Server-Side Apply
	deployToApply := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "app1",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app": "app1",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "nginx",
						},
					},
				},
			},
		},
	}
	deployToApply.SetName("nginx")
	deployToApply.SetNamespace("default")
	deployToApply.SetGroupVersionKind(
		appsv1.SchemeGroupVersion.WithKind("Deployment"),
	)
	err = a.client.Patch(
		ctx,
		&deployToApply,
		client.Apply,
		client.FieldOwner("mycontroller"),
		client.ForceOwnership,
	)
	if err != nil {
		fmt.Printf("patch: %v\n", err)
		return reconcile.Result{}, err
	}

	// ## Updating the status of a resource
	myresource.Status.State = "done"
	err = a.client.Status().Update(ctx, &myresource)
	if err != nil {
		fmt.Printf("status.update: %v\n", err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (a *MyReconciler) InjectClient(
	c client.Client,
) error {
	a.client = c
	return nil
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}
