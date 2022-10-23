package main

import (
	"context"
	"fmt"

	myresourcev1alpha1 "github.com/myid/myresource-crd/pkg/apis/mygroup.example.com/v1alpha1"
	"github.com/myid/myresource-crd/pkg/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	config, err :=
		clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			nil,
		).ClientConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	list, err := clientset.MygroupV1alpha1().
		MyResources("default").
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, res := range list.Items {
		fmt.Printf("%s\n", res.GetName())
	}

	// # Using the unstructured package and dynamic client
	// ## The Unstructured type
	u, err := getResource()
	if err != nil {
		panic(err)
	}

	// ## Converting between typed and unstructured objects
	converter := runtime.DefaultUnstructuredConverter

	// ### Unstructured to Typed
	var myresource myresourcev1alpha1.MyResource
	converter.FromUnstructured(
		u.UnstructuredContent(), &myresource,
	)

	fmt.Printf("%s, %s\n", myresource.Spec.Image, &myresource.Spec.Memory)

	// ### Typed to Unstructured
	var newU unstructured.Unstructured
	newU.Object, err = converter.ToUnstructured(&myresource)
	if err != nil {
		panic(err)
	}

	image, found, err := unstructured.NestedString(newU.Object, "spec", "image")
	if err != nil {
		panic(err)
	}
	if !found {
		panic("spec.image not found")
	}
	memory, found, err := unstructured.NestedString(newU.Object, "spec", "memory")
	if err != nil {
		panic(err)
	}
	if !found {
		panic("spec.memory not found")
	}
	fmt.Printf("%s, %s\n", image, memory)

	// # The dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	createdU, err := CreateMyResource(dynamicClient, u)
	if err != nil {
		panic(err)
	}
	_ = createdU
}
