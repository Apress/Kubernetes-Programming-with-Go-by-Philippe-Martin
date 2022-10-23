package main

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	mygroupv1alpha1 "github.com/myid/myresource-crd/pkg/apis/mygroup.example.com/v1alpha1"
)

func main() {
	log.SetLogger(zap.New())

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

	eventRecorder := mgr.GetEventRecorderFor(
		"MyResource",
	)

	err = builder.
		ControllerManagedBy(mgr).
		For(&mygroupv1alpha1.MyResource{}).
		Owns(&appsv1.Deployment{}).
		Complete(&MyReconciler{
			EventRecorder: eventRecorder,
		})
	panicIf(err)

	err = mgr.Start(context.Background())
	panicIf(err)
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}
