package main

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestCreateMyResourceWhenResourceExists(t *testing.T) {
	myres, err := getResource()
	if err != nil {
		t.Error(err)
	}

	dynamicClient := fake.NewSimpleDynamicClient(
		runtime.NewScheme(),
		myres,
	)

	// Not really used, just to show how to use it
	dynamicClient.Fake.PrependReactor(
		"create",
		"myresources",
		func(
			action ktesting.Action,
		) (handled bool, ret runtime.Object, err error) {
			return false, nil, nil
		})
	_, err = CreateMyResource(dynamicClient, myres)
	if err == nil {
		t.Error("Error should happen")
	}

	actions := dynamicClient.Fake.Actions()
	if len(actions) != 1 {
		t.Errorf("# of actions should be %d but is %d", 1, len(actions))
	}
}
