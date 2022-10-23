package main

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePod(
	ctx context.Context,
	clientset kubernetes.Interface,
	name string,
	namespace string,
	image string,
) (pod *corev1.Pod, err error) {

	podToCreate := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "runtime",
					Image: image,
				},
			},
		},
	}
	podToCreate.SetName(name)

	return clientset.CoreV1().
		Pods(namespace).
		Create(
			ctx,
			&podToCreate,
			metav1.CreateOptions{},
		)
}
