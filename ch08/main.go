package main

import (
	"context"
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := getConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	list, err := clientset.ApiextensionsV1().
		CustomResourceDefinitions().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, crd := range list.Items {
		fmt.Printf("%s\n", crd.GetName())
	}
}

func getConfig() (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
}
