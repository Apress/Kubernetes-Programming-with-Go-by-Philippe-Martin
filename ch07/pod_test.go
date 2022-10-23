package main

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

// TestCreatePod: Checking the result of the function
func TestCreatePod1(t *testing.T) {
	var (
		name      = "a-name"
		namespace = "a-namespace"
		image     = "an-image"

		wantPod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a-name",
				Namespace: "a-namespace",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "runtime",
						Image: "an-image",
					},
				},
			},
		}
	)

	clientset := fake.NewSimpleClientset()
	gotPod, err := CreatePod(
		context.Background(),
		clientset,
		name,
		namespace,
		image,
	)

	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
	if !reflect.DeepEqual(gotPod, wantPod) {
		t.Errorf("CreatePod() = %v, want %v",
			gotPod,
			wantPod,
		)
	}
}

// TestCreatePod2: Reacting to Actions
func TestCreatePod2(t *testing.T) {
	var (
		name      = "a-name"
		namespace = "a-namespace"
		image     = "an-image"

		wantPod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a-name",
				Namespace: "a-namespace",
			},
			Spec: corev1.PodSpec{
				NodeName: "node1",
				Containers: []corev1.Container{
					{
						Name:  "runtime",
						Image: "an-image",
					},
				},
			},
		}
	)

	clientset := fake.NewSimpleClientset()

	clientset.Fake.PrependReactor("create", "pods", func(
		action ktesting.Action,
	) (handled bool, ret runtime.Object, err error) {
		act := action.(ktesting.CreateAction)
		ret = act.GetObject()
		pod := ret.(*corev1.Pod)
		pod.Spec.NodeName = "node1"
		return false, pod, nil
	})

	gotPod, err := CreatePod(
		context.Background(),
		clientset,
		name,
		namespace,
		image,
	)

	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
	if !reflect.DeepEqual(gotPod, wantPod) {
		t.Errorf("CreatePod() = %v, want %v",
			gotPod,
			wantPod,
		)
	}
}

// TestCreatePod3: Checking the actions
func TestCreatePod3(t *testing.T) {
	var (
		name      = "a-name"
		namespace = "a-namespace"
		image     = "an-image"

		wantPod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "a-name",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "runtime",
						Image: "an-image",
					},
				},
			},
		}

		wantActions = 1
	)

	clientset := fake.NewSimpleClientset() // ➊
	_, _ = CreatePod(                      // ➋
		context.Background(),
		clientset,
		name,
		namespace,
		image,
	)

	actions := clientset.Actions()   // ➌
	if len(actions) != wantActions { // ➍
		t.Errorf("# actions = %d, want %d",
			len(actions),
			wantActions,
		)
	}
	action := actions[0] // ➎

	actionNamespace := action.GetNamespace() // ➏
	if actionNamespace != namespace {
		t.Errorf("action namespace = %s, want %s",
			actionNamespace,
			namespace,
		)
	}

	if !action.Matches("create", "pods") { // ➐
		t.Errorf("action verb = %s, want create",
			action.GetVerb(),
		)
		t.Errorf("action resource = %s, want pods",
			action.GetResource().Resource,
		)
	}

	createAction := action.(ktesting.CreateAction) // ➑
	obj := createAction.GetObject()                // ➒
	if !reflect.DeepEqual(obj, wantPod) {
		t.Errorf("create action object = %v, want %v",
			obj,
			wantPod,
		)
	}

}
