package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mygroupv1alpha1 "github.com/myid/myresource-crd/pkg/apis/mygroup.example.com/v1alpha1"
)

func main() {
	scheme := runtime.NewScheme() // ❶
	clientgoscheme.AddToScheme(scheme)
	mygroupv1alpha1.AddToScheme(scheme)

	mgr, err := manager.New( // ❷
		config.GetConfigOrDie(),
		manager.Options{
			Scheme: scheme,
		},
	)
	panicIf(err)

	controller, err := controller.New( // ❸
		"my-operator", mgr,
		controller.Options{
			Reconciler: &MyReconciler{},
		})
	panicIf(err)

	err = controller.Watch( // ❹
		&source.Kind{
			Type: &mygroupv1alpha1.MyResource{},
		},
		&handler.EnqueueRequestForObject{},
	)
	panicIf(err)

	err = controller.Watch( // ❺
		&source.Kind{
			Type: &corev1.Pod{},
		},
		&handler.EnqueueRequestForOwner{
			OwnerType:    &corev1.Pod{},
			IsController: true,
		},
	)
	panicIf(err)

	err = mgr.Start(context.Background()) // ❻
	panicIf(err)
}

type MyReconciler struct{} // ➐

func (o *MyReconciler) Reconcile( // ➑
	ctx context.Context,
	r reconcile.Request,
) (reconcile.Result, error) {
	fmt.Printf("reconcile %v\n", r)
	return reconcile.Result{}, nil
}

// panicIf panic if err is not nil
// Please call from main only!
func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}
