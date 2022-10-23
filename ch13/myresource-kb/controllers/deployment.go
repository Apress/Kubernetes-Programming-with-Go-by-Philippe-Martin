package controllers

import (
	"context"

	mygroupv1alpha1 "github.com/myid/myresource/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (a *MyResourceReconciler) applyDeployment(
	ctx context.Context,
	myres *mygroupv1alpha1.MyResource,
	ownerref *metav1.OwnerReference,
) error {
	deploy := createDeployment(myres, ownerref)
	err := a.Client.Patch(
		ctx,
		deploy,
		client.Apply,
		client.FieldOwner(Name),
		client.ForceOwnership,
	)
	//generation := deploy.GetGeneration()
	//if generation == 1 {
	//	a.EventRecorder.Eventf(myres, corev1.EventTypeNormal, "DeploymentCreated", "The deployment %q has been created", deploy.GetName())
	//}
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
							Image: myres.Spec.Image,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: myres.Spec.Memory,
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
	deploy.SetOwnerReferences([]metav1.OwnerReference{
		*ownerref,
	})
	return deploy
}
