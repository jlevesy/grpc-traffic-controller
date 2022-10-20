/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kxdsv1alpha1 "github.com/jlevesy/kxds/api/v1alpha1"
	"github.com/jlevesy/kxds/xds"
)

type Reconciller struct {
	client    client.Client
	refresher xds.Refresher
}

func NewReconciler(cl client.Client, refresher xds.Refresher) *Reconciller {
	return &Reconciller{
		client:    cl,
		refresher: refresher,
	}
}

//+kubebuilder:rbac:groups=api.kxds.dev,resources=services,verbs=get;list;watch;
//+kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		endpoints corev1.EndpointsList
		services  kxdsv1alpha1.ServiceList

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
