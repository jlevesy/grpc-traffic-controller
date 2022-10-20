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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kxdsv1alpha1 "github.com/jlevesy/kxds/api/v1alpha1"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client client.Client
}

func NewServiceReconciler(cl client.Client) *ServiceReconciler {
	return &ServiceReconciler{client: cl}
}

//+kubebuilder:rbac:groups=api.kxds.dev,resources=services,verbs=get;list;watch;
//+kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		svc          kxdsv1alpha1.Service
		svcEndpoints corev1.Endpoints

		logger = log.FromContext(ctx)
	)

	err := r.client.Get(ctx, req.NamespacedName, &svc)
	switch {
	case errors.IsNotFound(err):
		return ctrl.Result{}, nil
	case err != nil:
		return ctrl.Result{}, fmt.Errorf("could not find service: %w", err)
	default:
	}

	logger.Info("Updating routing configuration for service")

	err = r.client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: svc.Spec.Destination.Name}, &svcEndpoints)
	switch {
	case errors.IsNotFound(err):
		return ctrl.Result{}, nil
	case err != nil:
		return ctrl.Result{}, fmt.Errorf("could not find endpoints: %w", err)
	default:
	}

	if len(svcEndpoints.Subsets) == 0 {
		logger.Info("no endpoint subsets found, skipping update")
	}

	return ctrl.Result{}, nil
}
