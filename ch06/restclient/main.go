package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	ctx := context.Background()
	config, err := getConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// # Getting result as table
	restClient := clientset.CoreV1().RESTClient() // ➊
	req := restClient.Get().
		Namespace("project1"). // ➋
		Resource("pods").      // ➌
		SetHeader(             // ➍
			"Accept",
			fmt.Sprintf(
				"application/json;as=Table;v=%s;g=%s",
				metav1.SchemeGroupVersion.Version,
				metav1.GroupName,
			))

	var result metav1.Table // ➎
	err = req.Do(ctx).      // ➏
				Into(&result) // ➐
	if err != nil {
		panic(err)
	}

	for _, colDef := range result.ColumnDefinitions { // ➑
		// display header
		fmt.Printf("%v\t", colDef.Name)
	}
	fmt.Printf("\n")

	for _, row := range result.Rows { // ➒
		for _, cell := range row.Cells { // ➓
			// display cell
			fmt.Printf("%v\t", cell)
		}
		fmt.Printf("\n")
	}
}

func getConfig() (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
}
