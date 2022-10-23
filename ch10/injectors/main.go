package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/cache"
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
	cache  cache.Cache
	scheme *runtime.Scheme
}

func (a *MyReconciler) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	fmt.Printf("reconcile %v\n", req)
	fmt.Printf("client: %p\n", a.client)
	fmt.Printf("cache: %p\n", a.cache)
	fmt.Printf("scheme: %p\n", a.scheme)
	return reconcile.Result{}, nil
}

func (a *MyReconciler) InjectClient(
	c client.Client,
) error {
	a.client = c
	return nil
}

func (a *MyReconciler) InjectCache(
	c cache.Cache,
) error {
	a.cache = c
	return nil
}

func (a *MyReconciler) InjectScheme(
	s *runtime.Scheme,
) error {
	a.scheme = s
	return nil
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}
