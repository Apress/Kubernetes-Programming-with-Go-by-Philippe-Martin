package main

import (
	"context"
	"path/filepath"
	"testing"

	mygroupv1alpha1 "github.com/myid/myresource-crd/pkg/apis/mygroup.example.com/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func TestMyReconciler_Reconcile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t,
		"Controller Suite",
	)
}

var (
	testEnv   *envtest.Environment // ❶
	ctx       context.Context
	cancel    context.CancelFunc
	k8sClient client.Client // ❷
)

var _ = BeforeSuite(func() {
	log.SetLogger(zap.New(
		zap.WriteTo(GinkgoWriter),
		zap.UseDevMode(true),
	))

	ctx, cancel = context.WithCancel( // ❸
		context.Background(),
	)

	testEnv = &envtest.Environment{ // ❹
		CRDDirectoryPaths: []string{
			filepath.Join("crd"),
		},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err := testEnv.Start() // ❺
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	scheme := runtime.NewScheme() // ❻
	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = mygroupv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	mgr, err := manager.New(cfg, manager.Options{ // ❼
		Scheme: scheme,
	})
	Expect(err).ToNot(HaveOccurred())
	k8sClient = mgr.GetClient() // ❽

	err = builder. // ❾
			ControllerManagedBy(mgr).
			Named(Name).
			For(&mygroupv1alpha1.MyResource{}).
			Owns(&appsv1.Deployment{}).
			Complete(&MyReconciler{})

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx) // ❿
		Expect(err).ToNot(
			HaveOccurred(),
			"failed to run manager",
		)
	}()
})

var _ = AfterSuite(func() {
	cancel()              // ⓫
	err := testEnv.Stop() // ⓬
	Expect(err).NotTo(HaveOccurred())
})
