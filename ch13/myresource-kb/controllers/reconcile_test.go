package controllers

import (
	"fmt"
	"math/rand"

	mygroupv1alpha1 "github.com/myid/myresource/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("MyResource controller", func() {
	When("When creating a MyResource instance", func() {

		var (
			myres      mygroupv1alpha1.MyResource
			ownerref   *metav1.OwnerReference
			name       string
			namespace  = "default"
			deployName string
			image      string
		)

		BeforeEach(func() {
			image = fmt.Sprintf("myimage%d", rand.Intn(1000))
			myres = mygroupv1alpha1.MyResource{
				Spec: mygroupv1alpha1.MyResourceSpec{
					Image: image,
				},
			}
			name = fmt.Sprintf("myres%d", rand.Intn(1000))
			myres.SetName(name)
			myres.SetNamespace(namespace)
			err := k8sClient.Create(ctx, &myres)
			Expect(err).NotTo(HaveOccurred())
			ownerref = metav1.NewControllerRef(
				&myres,
				mygroupv1alpha1.GroupVersion.WithKind("MyResource"),
			)
			deployName = fmt.Sprintf("%s-deployment", name)
		})

		AfterEach(func() {
			k8sClient.Delete(ctx, &myres)
		})

		It("should create a deployment", func() {
			var dep appsv1.Deployment
			Eventually(deploymentExists(deployName, namespace, &dep), 10, 1).
				Should(BeTrue())
		})

		When("deployment is found", func() {
			var dep appsv1.Deployment

			BeforeEach(func() {
				Eventually(
					deploymentExists(deployName, namespace, &dep), 10, 1,
				).Should(BeTrue())
			})

			It("should be owned by the MyResource instance", func() {
				Expect(dep.GetOwnerReferences()).
					To(ContainElement(*ownerref))
			})

			It("should use the image specified in MyResource instance", func() {
				Expect(dep.Spec.Template.Spec.Containers[0].Image).
					To(Equal(image))
			})

			When("deployment ReadyReplicas is 1", func() {
				BeforeEach(func() {
					dep.Status.Replicas = 1
					dep.Status.ReadyReplicas = 1
					err := k8sClient.Status().Update(ctx, &dep)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should set status ready for MyResource instance", func() {
					Eventually(getMyResourceState(name, namespace), 10, 1).
						Should(Equal("Ready"))
				})
			})
		})
	})
})

func deploymentExists(name, namespace string, dep *appsv1.Deployment) func() bool {
	return func() bool {
		err := k8sClient.Get(ctx, client.ObjectKey{
			Namespace: namespace,
			Name:      name,
		}, dep)
		return err == nil
	}
}

func getMyResourceState(name, namespace string) func() (string, error) {
	return func() (string, error) {
		myres := mygroupv1alpha1.MyResource{}
		err := k8sClient.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}, &myres)
		if err != nil {
			return "", err
		}
		return myres.Status.State, nil
	}
}
