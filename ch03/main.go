package main

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// # Specific content in core/v1
	// ## ResourceList
	requests := corev1.ResourceList{
		corev1.ResourceMemory: *resource.NewQuantity(64*1024*1024, resource.BinarySI),
		corev1.ResourceCPU:    *resource.NewMilliQuantity(250, resource.DecimalSI),
	}
	_ = requests

	// # Writing Kubernetes Resources in Go
	// ## Importing the package
	myDep := appsv1.Deployment{}
	_ = myDep

	// ## The ObjectMeta fields
	// ### Name
	configmap := corev1.ConfigMap{}
	configmap.SetName("config")

	// ### Labels and annotations
	mylabels := map[string]string{
		"app.kubernetes.io/component": "my-component",
		"app.kubernetes.io/name":      "a-name",
	}

	mylabels["app.kubernetes.io/part-of"] = "my-app"

	mySet := labels.Set{
		"app.kubernetes.io/component": "my-component",
		"app.kubernetes.io/name":      "a-name",
	}
	mySet["app.kubernetes.io/part-of"] = "my-app"

	// ### OwnerReferences
	// Get the object to reference
	clientset, err := getClientset()
	if err != nil {
		panic(err)
	}
	pod, err := clientset.CoreV1().Pods("myns").
		Get(context.TODO(), "mypodname", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	// Solution 1: set the APIVersion and Kind of the Pod
	// then copy all information from the pod

	pod.SetGroupVersionKind(
		corev1.SchemeGroupVersion.WithKind("Pod"),
	)
	ownerRef := metav1.OwnerReference{
		APIVersion: pod.APIVersion,
		Kind:       pod.Kind,
		Name:       pod.GetName(),
		UID:        pod.GetUID(),
	}

	// Solution 2: Copy name and uid from pod
	// then set APIVersion and Kind on the OwnerReference

	ownerRef = metav1.OwnerReference{
		Name: pod.GetName(),
		UID:  pod.GetUID(),
	}
	ownerRef.APIVersion, ownerRef.Kind =
		corev1.SchemeGroupVersion.WithKind("Pod").
			ToAPIVersionAndKind()

	// #### Setting Controller
	// Solution 1: declare a value and use its address
	controller := true
	ownerRef.Controller = &controller

	// Solution 2: use the BoolPtr function
	ownerRef.Controller = pointer.BoolPtr(true)

	// ## Comparison with writing YAML manifests
	// Solution 1
	pod1 := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "runtime",
					Image: "nginx",
				},
			},
		},
	}
	pod1.SetName("my-pod")
	pod1.SetLabels(map[string]string{
		"component": "my-component",
	})

	// Solution 2
	pod2 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx",
			Labels: map[string]string{
				"component": "mycomponent",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "runtime",
					Image: "nginx",
				},
			},
		},
	}
	_ = pod2
}

func getClientset() (*kubernetes.Clientset, error) {
	config, err :=
		clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			nil,
		).ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
