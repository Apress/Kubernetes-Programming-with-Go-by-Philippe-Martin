package main

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

func getPods(
	ctx context.Context,
	restClient rest.Interface,
	ns string,
) ([]corev1.Pod, error) {
	result := corev1.PodList{}
	err := restClient.Get().
		Namespace(ns).
		Resource("pods").
		Do(ctx).
		Into(&result)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}
