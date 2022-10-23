package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	// # Connecting to the cluster
	config, err := getConfig5()
	if err != nil {
		panic(err)
	}

	// # Examining the requests
	klog.InitFlags(nil)
	flag.Parse()

	// # Getting a clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// # Creating a resource
	ctx := context.Background()
	wantedPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"mykey": "value1",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx",
				},
			},
		},
	}
	wantedPod.SetName("nginx-pod")

	createdPod, err := clientset.
		CoreV1().
		Pods("project1").
		Create(ctx, &wantedPod, metav1.CreateOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("Namespace %q not found\n", "project1")
			os.Exit(1)
		} else if errors.IsAlreadyExists(err) {
			fmt.Printf("Pod %q already exists\n", "nginx-pod")
			os.Exit(1)
		} else if errors.IsInvalid(err) {
			fmt.Printf("Pod specification is invalid\n")
			os.Exit(1)
		}
		panic(err)
	}
	_ = createdPod

	// # Getting information about a resource
	pod, err := clientset.
		CoreV1().
		Pods("project1").
		Get(ctx, "nginx-pod", metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("Pod %q is not found\n", "nginx-pod")
			os.Exit(1)
		}
		panic(err)
	}
	_ = pod

	// # Getting list of resources
	podListInNS, err := clientset.
		CoreV1().
		Pods("project1").
		List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, pod := range podListInNS.Items {
		fmt.Printf("%s\n", pod.GetName())
	}

	podListInCluster, err := clientset.
		CoreV1().
		Pods("").
		List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, pod := range podListInCluster.Items {
		fmt.Printf("%s\n", pod.GetName())
	}

	// # Filtering the result of a list
	// ## Setting LabelSelector using the labels package
	// ### Using Requirements
	req1, err := labels.NewRequirement(
		"mykey",
		selection.Equals,
		[]string{"value1"},
	)
	if err != nil {
		panic(err)
	}
	labelsSelector := labels.NewSelector()
	labelsSelector = labelsSelector.Add(*req1)
	s := labelsSelector.String()
	filteredPodListInCluster, err := clientset.
		CoreV1().
		Pods("").
		List(ctx, metav1.ListOptions{
			LabelSelector: s,
		})
	if err != nil {
		panic(err)
	}
	for _, pod := range filteredPodListInCluster.Items {
		fmt.Printf("%s\n", pod.GetName())
	}

	// ### Parsing a LabelSelector string
	selector, err := labels.Parse(
		"mykey = value1",
	)
	if err != nil {
		panic(err)
	}
	s = selector.String()
	filteredPodListInCluster, err = clientset.
		CoreV1().
		Pods("").
		List(ctx, metav1.ListOptions{
			LabelSelector: s,
		})
	if err != nil {
		panic(err)
	}
	for _, pod := range filteredPodListInCluster.Items {
		fmt.Printf("%s\n", pod.GetName())
	}

	// ### Using a set of key-value pairs
	set := labels.Set{
		"mykey": "value1",
	}
	selector, err = labels.ValidatedSelectorFromSet(set)
	if err != nil {
		panic(err)
	}
	s = selector.String()
	filteredPodListInCluster, err = clientset.
		CoreV1().
		Pods("").
		List(ctx, metav1.ListOptions{
			LabelSelector: s,
		})
	if err != nil {
		panic(err)
	}
	for _, pod := range filteredPodListInCluster.Items {
		fmt.Printf("%s\n", pod.GetName())
	}

	// ## Setting Fieldselector using the fields package
	// ### Assembling one term selectors
	fselector := fields.AndSelectors(
		fields.OneTermEqualSelector(
			"status.phase",
			"Running",
		),
		fields.OneTermNotEqualSelector(
			"spec.restartPolicy",
			"Always",
		),
	)
	fs := fselector.String()
	filteredPodListInCluster, err = clientset.
		CoreV1().
		Pods("").
		List(ctx, metav1.ListOptions{
			FieldSelector: fs,
		})
	if err != nil {
		panic(err)
	}
	for _, pod := range filteredPodListInCluster.Items {
		fmt.Printf("%s\n", pod.GetName())
	}

	// ### Parsing a FieldSelector string
	sel, err := fields.ParseSelector(
		"status.phase=Running,spec.restartPolicy!=Always",
	)
	if err != nil {
		panic(err)
	}
	fs = sel.String()
	filteredPodListInCluster, err = clientset.
		CoreV1().
		Pods("").
		List(ctx, metav1.ListOptions{
			FieldSelector: fs,
		})
	if err != nil {
		panic(err)
	}
	for _, pod := range filteredPodListInCluster.Items {
		fmt.Printf("%s\n", pod.GetName())
	}

	// ### Using a set of key-value pairs
	fset := fields.Set{
		"status.phase":       "Running",
		"spec.restartPolicy": "Never",
	}

	sel = fields.SelectorFromSet(fset)
	fs = sel.String()
	filteredPodListInCluster, err = clientset.
		CoreV1().
		Pods("").
		List(ctx, metav1.ListOptions{
			FieldSelector: fs,
		})
	if err != nil {
		panic(err)
	}
	for _, pod := range filteredPodListInCluster.Items {
		fmt.Printf("%s\n", pod.GetName())
	}

	uid := createdPod.GetUID()
	rv := createdPod.GetResourceVersion()

	// # Deleting a resource
	err = clientset.
		CoreV1().
		Pods("project1").
		Delete(ctx, "nginx-pod", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}

	// ## With grace period
	err = clientset.
		CoreV1().
		Pods("project1").
		Delete(ctx, "nginx-pod", *metav1.NewDeleteOptions(5))
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("pod %q already deleted\n", "nginx-pod")
		} else {
			panic(err)
		}
	}

	// ## Using UID precondition
	err = clientset.
		CoreV1().
		Pods("project1").
		Delete(ctx, "nginx-pod", *metav1.NewPreconditionDeleteOptions(
			string(uid),
		))
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("pod %q already deleted\n", "nginx-pod")
		} else if errors.IsConflict(err) {
			fmt.Printf("Conflicting UID %q\n", string(uid))
		} else {
			panic(err)
		}
	}

	// ## Using ResourceVersion precondition
	err = clientset.
		CoreV1().
		Pods("project1").
		Delete(ctx, "nginx-pod", *metav1.NewRVDeletionPrecondition(
			rv,
		))
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("pod %q already deleted\n", "nginx-pod")
		} else if errors.IsConflict(err) {
			// This error will be raised, as the resource has changed due to previous deletion
			fmt.Printf("Conflicting resourceVersion %q\n", string(rv))
		} else {
			panic(err)
		}
	}

	// ## With Propagation policy
	options := *metav1.NewDeleteOptions(5)
	policy := metav1.DeletePropagationForeground
	options.PropagationPolicy = &policy
	err = clientset.
		CoreV1().
		Pods("project1").
		Delete(ctx, "nginx-pod", options)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("pod %q already deleted\n", "nginx-pod")
		} else {
			panic(err)
		}
	}

	// # Deleting a collection of resources
	err = clientset.
		CoreV1().
		Pods("project1").
		DeleteCollection(
			ctx,
			metav1.DeleteOptions{},
			metav1.ListOptions{},
		)
	if err != nil {
		panic(err)
	}

	// # Updating a resource
	wantedDep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "app1",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "app1",
					},
				},
				Spec: v1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
						},
					},
				},
			},
		},
	}
	wantedDep.SetName("nginx")

	createdDep, err := clientset.
		AppsV1().
		Deployments("project1").
		Create(ctx, &wantedDep, metav1.CreateOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("Namespace %q not found\n", "project1")
			os.Exit(1)
		} else if errors.IsAlreadyExists(err) {
			fmt.Printf("Deployment %q already exists\n", "nginx-pod")
			os.Exit(1)
		} else if errors.IsInvalid(err) {
			fmt.Printf("Deployment specification is invalid: %v\n", err)
			os.Exit(1)
		}
		panic(err)
	}
	_ = createdDep

	wantedDep.Spec.Template.Spec.Containers[0].Name = "main"
	updatedDep, err := clientset.
		AppsV1().
		Deployments("project1").
		Update(
			ctx,
			&wantedDep,
			metav1.UpdateOptions{},
		)
	if err != nil {
		if errors.IsInvalid(err) {
			fmt.Printf("Deployment specification is invalid: %v\n", err)
			os.Exit(1)
		} else if errors.IsConflict(err) {
			fmt.Printf("Conflict updating deployment %q\n", "nginx")
			os.Exit(1)
		}
		panic(err)
	}

	time.Sleep(3 * time.Second)

	// # Using a strategic merge patch to update a resource
	for {
		// We want to retry as the Deployment will be modified
		// by the controller to add default fields
		// and we need to read and patch the stabilized version
		conflict := false

		existingDep, err := clientset.
			AppsV1().
			Deployments("project1").
			Get(ctx, "nginx", metav1.GetOptions{})

		if err != nil {
			if errors.IsNotFound(err) {
				fmt.Printf("Deployment %q is not found\n", "nginx")
				os.Exit(1)
			}
			panic(err)
		}

		patch := client.StrategicMergeFrom(
			existingDep,
			client.MergeFromWithOptimisticLock{},
		)
		updatedDep2 := updatedDep.DeepCopy()
		updatedDep2.Spec.Replicas = pointer.Int32(2)
		patchData, err := patch.Data(updatedDep2)
		if err != nil {
			panic(err)
		}
		patchedDep, err := clientset.
			AppsV1().Deployments("project1").Patch(
			ctx,
			"nginx",
			patch.Type(),
			patchData,
			metav1.PatchOptions{},
		)
		if err != nil {
			if errors.IsInvalid(err) {
				fmt.Printf("Deployment specification is invalid: %v\n", err)
				os.Exit(1)
			} else if errors.IsConflict(err) {
				fmt.Printf("Conflict patching deployment %q: %v\nRetrying...\n", "nginx", err)
				conflict = true
			} else {
				panic(err)
			}
		}
		_ = patchedDep
		if !conflict {
			break
		}
		time.Sleep(1 * time.Second)
	}

	time.Sleep(3 * time.Second)

	// # Applying resources server-side with Patch
	ssaDep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			// there is no way to indicate a default value for this field,
			// the nil value being valid,
			// so we need to indicate it
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "app1",
				},
			},
			Replicas: pointer.Int32(1), // This value is changed
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					// there is no way to indicate a default value for this field,
					// the nil value being valid,
					// so we need to indicate it
					Labels: map[string]string{
						"app": "app1",
					},
				},
			},
		},
	}
	ssaDep.SetName("nginx")

	ssaDep.APIVersion, ssaDep.Kind =
		appsv1.SchemeGroupVersion.
			WithKind("Deployment").
			ToAPIVersionAndKind()

	patch := client.Apply
	patchData, err := patch.Data(&ssaDep)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(patchData))

	patchedDep, err := clientset.
		AppsV1().Deployments("project1").Patch(
		ctx,
		"nginx",
		patch.Type(),
		patchData,
		metav1.PatchOptions{
			FieldManager: "my-program",
			Force:        pointer.Bool(true),
		},
	)
	if err != nil {
		if errors.IsInvalid(err) {
			fmt.Printf("Deployment specification is invalid: %v\n", err)
			os.Exit(1)
		} else if errors.IsConflict(err) {
			fmt.Printf("Conflict server-side patching deployment %q: %v\n", "nginx", err)
			os.Exit(1)
		} else {
			panic(err)
		}
	}
	_ = patchedDep

	time.Sleep(3 * time.Second)

	// # Server-Side Apply using Apply Configurations
	deployConfig := acappsv1.Deployment(
		"nginx",
		"project1",
	)
	deployConfig.WithSpec(acappsv1.DeploymentSpec())
	deployConfig.Spec.WithReplicas(2)

	patchedDep, err = clientset.AppsV1().
		Deployments("project1").Apply(
		ctx,
		deployConfig,
		metav1.ApplyOptions{
			FieldManager: "my-program",
			Force:        true,
		},
	)
	if err != nil {
		if errors.IsInvalid(err) {
			fmt.Printf("Deployment specification is invalid: %v\n", err)
			os.Exit(1)
		} else if errors.IsConflict(err) {
			fmt.Printf("Conflict server-side applying deployment %q: %v\n", "nginx", err)
			os.Exit(1)
		} else {
			panic(err)
		}
	}
	_ = patchedDep

	// # Watching resources
	watcher, err := clientset.AppsV1().
		Deployments("project1").
		Watch(
			ctx,
			metav1.ListOptions{},
		)
	if err != nil {
		panic(err)
	}

	fmt.Printf("==============================\nWatching, press Ctrl-c to exit\n==============================\n")
	for ev := range watcher.ResultChan() {
		switch v := ev.Object.(type) {
		case *appsv1.Deployment:
			fmt.Printf("%s %s\n", ev.Type, v.GetName())
		case *metav1.Status:
			fmt.Printf("%s\n", v.Status)
			watcher.Stop()
		}
	}

}

// # Connecting to the cluster

// ## In-cluster configuration
func getConfig1() (*rest.Config, error) {
	return rest.InClusterConfig()
}

// ## Out-of-cluster configuration
// ### From kubeconfig in memory
func getConfig2() (*rest.Config, error) {
	configBytes, err := os.ReadFile(
		"/home/user/.kube/config",
	)
	if err != nil {
		return nil, err
	}
	return clientcmd.RESTConfigFromKubeConfig(
		configBytes,
	)
}

// ### From a kubeconfig on disk
func getConfig3() (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags(
		"",
		"/home/user/.kube/config",
	)
}

// ### From a personalized kubeconfig
func getConfig4() (*rest.Config, error) {
	return clientcmd.BuildConfigFromKubeconfigGetter(
		"",
		func() (*api.Config, error) {
			apiConfig, err := clientcmd.LoadFromFile(
				"/home/user/.kube/config",
			)
			if err != nil {
				return nil, nil
			}
			// TODO: manipulate apiConfig
			return apiConfig, nil
		},
	)
}

// ### From several kubeconfig files
func getConfig5() (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
}

// ### Overriding kubeconfig with CLI flags
func getConfig6() (*rest.Config, error) {
	var (
		flags     pflag.FlagSet
		overrides clientcmd.ConfigOverrides
		of        = clientcmd.RecommendedConfigOverrideFlags("")
	)
	clientcmd.BindOverrideFlags(&overrides, &flags, of)
	flags.Parse(os.Args[1:])

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&overrides,
	).ClientConfig()

}
