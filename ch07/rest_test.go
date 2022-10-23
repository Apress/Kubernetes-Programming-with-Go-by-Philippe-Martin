package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest/fake"
)

func Test_getPods1(t *testing.T) {
	restClient := &fake.RESTClient{
		GroupVersion:         corev1.SchemeGroupVersion,
		NegotiatedSerializer: scheme.Codecs,

		Err: errors.New("an error from the rest client"),
	}

	_, err := getPods(
		context.Background(),
		restClient,
		"default",
	)

	status, ok := err.(*url.Error)
	if !ok {
		t.Errorf("err should be of type url.Error")
	}
	if status.Err.Error() != errors.New(`an error from the rest client`).Error() {
		t.Errorf("Error is %v\n", status.Err.Error())
	}
}

func Test_getPods2(t *testing.T) {
	restClient := &fake.RESTClient{
		GroupVersion:         corev1.SchemeGroupVersion,
		NegotiatedSerializer: scheme.Codecs,

		Err: nil,
		Resp: &http.Response{
			StatusCode: http.StatusNotFound,
		},
	}

	_, err := getPods(
		context.Background(),
		restClient,
		"default",
	)

	status, ok := err.(*kerrors.StatusError)
	if !ok {
		t.Errorf("err should be of type errors.StatusError")
	}
	code := status.Status().Code
	if code != http.StatusNotFound {
		t.Errorf("Error code must be %d but is %d\n", http.StatusNotFound, code)
	}
}
