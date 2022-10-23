package main

import (
	"bytes"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
)

func main() {
	// # Scheme
	// ## Initialization
	scheme1 := runtime.NewScheme()
	scheme1.AddKnownTypes(
		schema.GroupVersion{
			Group:   "",
			Version: "v1",
		},
		&corev1.Pod{},
		&corev1.ConfigMap{},
	)

	scheme2 := runtime.NewScheme()
	scheme2.AddKnownTypes(
		schema.GroupVersion{
			Group:   "apps",
			Version: "v1",
		},
		&appsv1.Deployment{},
	)
	scheme2.AddKnownTypes(
		schema.GroupVersion{
			Group:   "apps",
			Version: "v1beta1",
		},
		&appsv1beta1.Deployment{},
	)

	// ## Mapping
	types := scheme2.KnownTypes(schema.GroupVersion{
		Group:   "apps",
		Version: "v1",
	})
	fmt.Printf("known types for scheme2: %v\n", types)
	// map[Deployment:v1.Deployment]

	groupVersions := scheme2.VersionsForGroupKind(
		schema.GroupKind{
			Group: "apps",
			Kind:  "Deployment",
		})
	fmt.Printf("groupVersions for scheme2: %v\n", groupVersions)
	// [apps/v1 apps/v1beta1]

	gvks, _, _ := scheme2.ObjectKinds(&appsv1.Deployment{})
	fmt.Printf("gvks for scheme2: %v\n", gvks)
	// [apps/v1, Kind=Deployment]

	obj, _ := scheme2.New(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
	fmt.Printf("obj: %v\n", obj)
	// &Deployment{ObjectMeta:{...} ...}

	// ## Conversion

	// ### Adding Conversion functions
	scheme2.AddConversionFunc(
		(*appsv1.Deployment)(nil),
		(*appsv1beta1.Deployment)(nil),
		func(a, b interface{}, scope conversion.Scope) error {
			v1deploy := a.(*appsv1.Deployment)
			v1beta1deploy := b.(*appsv1beta1.Deployment)
			// make conversion here
			_ = v1deploy
			_ = v1beta1deploy
			return nil
		})

	// ### Converting
	v1deployment := appsv1.Deployment{}
	v1deployment.SetName("myname")
	v1deployment.APIVersion, v1deployment.Kind =
		appsv1.SchemeGroupVersion.WithKind("Deployment").ToAPIVersionAndKind()

	var v1beta1Deployment appsv1beta1.Deployment
	scheme2.Convert(&v1deployment, &v1beta1Deployment, nil)

	// ## Serialization
	// ### JSON and YAML jsonSerializer
	jsonSerializer := jsonserializer.NewSerializerWithOptions(
		jsonserializer.SimpleMetaFactory{},
		scheme2,
		scheme2,
		jsonserializer.SerializerOptions{
			Yaml:   false, // or true for YAML serializer
			Pretty: true,  // or false for one-line JSON
			Strict: false, // or true to check duplicates
		},
	)

	// ### Protobuf serializer
	pbSerializer := protobuf.NewSerializer(scheme2, scheme2)
	_ = pbSerializer

	// ### Encoding and Decoding
	var buffer bytes.Buffer
	jsonSerializer.Encode(&v1deployment, &buffer)
	fmt.Printf("%s\n", buffer.String())
	// {
	//	"kind": "Deployment",
	//	"apiVersion": "apps/v1",
	// 	"metadata": {
	// 	  "name": "myname",
	// 	  "creationTimestamp": null
	// 	},
	//  ...

	var decodedDeployment appsv1.Deployment
	json := `{"kind": "Deployment", "apiVersion": "apps/v1", "metadata":{"name":"myname"}}`
	obj, gvk, _ := jsonSerializer.Decode(
		[]byte(json),
		nil,
		&decodedDeployment,
	)
	fmt.Printf("obj: %v\ngvk: %s\n", decodedDeployment, gvk)
	// obj: {{Deployment apps/v1} {myname ...
	// gvk: apps/v1, Kind=Deployment

	// # RESTMapper

	restMapper := meta.NewDefaultRESTMapper(groupVersions)
	restMapper.Add(appsv1beta1.SchemeGroupVersion.WithKind("Deployment"), nil)
	restMapper.Add(appsv1.SchemeGroupVersion.WithKind("Deployment"), nil)

	//	restMapper = meta.NewDefaultRESTMapper(groupVersions)
	//	restMapper.AddSpecific(
	//		appsv1.SchemeGroupVersion.WithKind("Deployment"),
	//		appsv1.SchemeGroupVersion.WithResource("deployments"),
	//		appsv1.SchemeGroupVersion.WithResource("deployment"),
	//		nil,
	//	)

	mapping, _ := restMapper.RESTMapping(
		schema.GroupKind{Group: "apps", Kind: "Deployment"},
	)
	fmt.Printf("single mapping: %+v\n", *mapping)
	// {Resource:apps/v1, Resource=deployments GroupVersionKind:apps/v1, Kind=Deployment Scope:<nil>}

	mappings, _ := restMapper.RESTMappings(
		schema.GroupKind{Group: "apps", Kind: "Deployment"},
	)
	for _, mapping := range mappings {
		fmt.Printf("mapping: %+v\n", *mapping)
	}
	// {Resource:apps/v1, Resource=deployments GroupVersionKind:apps/v1, Kind=Deployment Scope:<nil>}
	// {Resource:apps/v1beta1, Resource=deployments GroupVersionKind:apps/v1beta1, Kind=Deployment Scope:<nil>}

	kinds, _ := restMapper.KindsFor(
		schema.GroupVersionResource{Group: "", Version: "", Resource: "deployment"},
	)
	fmt.Printf("kinds: %+v\n", kinds)
	// [apps/v1, Kind=Deployment apps/v1beta1, Kind=Deployment]

	resources, _ := restMapper.ResourcesFor(
		schema.GroupVersionResource{Group: "", Version: "", Resource: "deployment"},
	)
	fmt.Printf("resources: %+v\n", resources)
	// [apps/v1, Resource=deployments apps/v1beta1, Resource=deployments]

}
