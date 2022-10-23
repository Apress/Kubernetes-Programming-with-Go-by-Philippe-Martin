package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func main() {
	err := createPod()
	if err != nil {
		panic(err)
	}
}

func createPod() error {
	pod := createPodObject() // ➊

	serializer := getJSONSerializer()
	postBody, err := serializePodObject(serializer, pod) // ➋
	if err != nil {
		return err
	}

	reqCreate, err := buildPostRequest(postBody) // ➌
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(reqCreate) // ➍
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body) // ➎
	if err != nil {
		return err
	}

	if resp.StatusCode < 300 { // ➏
		createdPod, err := deserializePodBody(serializer, body) // ➐
		if err != nil {
			return err
		}
		json, err := json.MarshalIndent(createdPod, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", json) // ➑
	} else {
		status, err := deserializeStatusBody(serializer, body) // ➒
		if err != nil {
			return err
		}
		json, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", json) // ➓
	}
	return nil
}

func createPodObject() *corev1.Pod { // ➊
	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "runtime",
					Image: "nginx",
				},
			},
		},
	}

	pod.SetName("my-pod")
	pod.SetLabels(map[string]string{
		"app.kubernetes.io/component": "my-component",
		"app.kubernetes.io/name":      "a-name",
	})
	return &pod
}

func serializePodObject( // ➋
	serializer runtime.Serializer,
	pod *corev1.Pod,
) (
	io.Reader,
	error,
) {
	var buf bytes.Buffer
	err := serializer.Encode(pod, &buf)
	if err != nil {
		return nil, err
	}
	return &buf, nil
}

func buildPostRequest( // ➌
	body io.Reader,
) (
	*http.Request,
	error,
) {
	reqCreate, err := http.NewRequest(
		"POST",
		"http://127.0.0.1:8001/api/v1/namespaces/default/pods",
		body,
	)
	if err != nil {
		return nil, err
	}
	reqCreate.Header.Add(
		"Accept",
		"application/json",
	)
	reqCreate.Header.Add(
		"Content-Type",
		"application/json",
	)
	return reqCreate, nil
}

func deserializePodBody( // ➐
	serializer runtime.Serializer,
	body []byte,
) (
	*corev1.Pod,
	error,
) {
	var result corev1.Pod
	_, _, err := serializer.Decode(body, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func deserializeStatusBody( // ➒
	serializer runtime.Serializer,
	body []byte,
) (
	*metav1.Status,
	error,
) {
	var status metav1.Status
	_, _, err := serializer.Decode(body, nil, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func getJSONSerializer() runtime.Serializer {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(
		schema.GroupVersion{
			Group:   "",
			Version: "v1",
		},
		&corev1.Pod{},
		&metav1.Status{},
	)
	return kjson.NewSerializerWithOptions(
		kjson.SimpleMetaFactory{},
		nil,
		scheme,
		kjson.SerializerOptions{},
	)
}
