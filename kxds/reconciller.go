package kxds

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kxdsv1alpha1 "github.com/jlevesy/kxds/api/v1alpha1"
)

type Reconciller struct {
	client    client.Client
	refresher Refresher
}

func NewReconciler(cl client.Client, refresher Refresher) *Reconciller {
	return &Reconciller{
		client:    cl,
		refresher: refresher,
	}
}

//+kubebuilder:rbac:groups=api.kxds.dev,resources=xdsservices,verbs=get;list;watch;
//+kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		endpoints corev1.EndpointsList
		services  kxdsv1alpha1.XDSServiceList

		logger = log.FromContext(ctx)
	)

	if err := r.client.List(ctx, &endpoints); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not gather endpoints list %w", err)
	}

	if err := r.client.List(ctx, &services); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not gather services list %w", err)
	}

	logger.Info("Triggering a cache refresh")

	return ctrl.Result{}, r.refresher.RefreshCache(ctx, services.Items, mapEndpointsByName(endpoints.Items))
}

func mapEndpointsByName(items []corev1.Endpoints) map[types.NamespacedName]corev1.Endpoints {
	result := make(map[types.NamespacedName]corev1.Endpoints, len(items))

	for _, i := range items {
		result[types.NamespacedName{Name: i.Name, Namespace: i.Namespace}] = i
	}

	return result
}
