package main

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := getConfig()
	if err != nil {
		panic(err)
	}

	client, err := discovery.NewDiscoveryClientForConfig(config)

	restMapper :=
		restmapper.NewDeferredDiscoveryRESTMapper(
			memory.NewMemCacheClient(client),
		)
	mapping, _ := restMapper.RESTMapping(
		schema.GroupKind{Group: "apps", Kind: "Deployment"},
	)
	fmt.Printf("single mapping: %+v\n", *mapping)
	// {Resource:apps/v1, Resource=deployments GroupVersionKind:apps/v1, Kind=Deployment Scope:0x1e68d00}

	mappings, _ := restMapper.RESTMappings(
		schema.GroupKind{Group: "apps", Kind: "Deployment"},
	)
	for _, mapping := range mappings {
		fmt.Printf("mapping: %+v\n", *mapping)
	}
	// {Resource:apps/v1, Resource=deployments GroupVersionKind:apps/v1, Kind=Deployment Scope:0x1e68d00}

	kinds, _ := restMapper.KindsFor(
		schema.GroupVersionResource{Group: "", Version: "", Resource: "deployment"},
	)
	fmt.Printf("kinds: %+v\n", kinds)
	// [apps/v1, Kind=Deployment]

	resources, _ := restMapper.ResourcesFor(
		schema.GroupVersionResource{Group: "", Version: "", Resource: "deployment"},
	)
	fmt.Printf("resources: %+v\n", resources)
	// [apps/v1, Resource=deployments]
}

func getConfig() (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
}
