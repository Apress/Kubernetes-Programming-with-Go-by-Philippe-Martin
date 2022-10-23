package controllers

import (
	"context"
	"fmt"

	mygroupv1alpha1 "github.com/myid/myresource/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	_buildingState = "Building"
	_readyState    = "Ready"
)

func (a *MyResourceReconciler) computeStatus(
	ctx context.Context,
	myres *mygroupv1alpha1.MyResource,
) (*mygroupv1alpha1.MyResourceStatus, error) {

	logger := log.FromContext(ctx)
	result := mygroupv1alpha1.MyResourceStatus{
		State: _buildingState,
	}

	deployList := appsv1.DeploymentList{}
	err := a.Client.List(
		ctx,
		&deployList,
		client.InNamespace(myres.GetNamespace()),
		client.MatchingLabels{
			"myresource": myres.GetName(),
		},
	)
	if err != nil {
		return nil, err
	}

	if len(deployList.Items) == 0 {
		logger.Info("no deployment found")
		return &result, nil
	}

	if len(deployList.Items) > 1 {
		logger.Info("too many deployments found", "count",
			len(deployList.Items))
		return nil, fmt.Errorf("%d deployment found, expected 1",
			len(deployList.Items))
	}

	status := deployList.Items[0].Status
	logger.Info("got deployment status", "status", status)
	if status.ReadyReplicas == 1 {
		result.State = _readyState
	}

	return &result, nil
}
