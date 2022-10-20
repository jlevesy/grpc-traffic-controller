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

package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	kxdsv1alpha1 "github.com/jlevesy/kxds/api/v1alpha1"
	"github.com/jlevesy/kxds/controllers"
	"github.com/jlevesy/kxds/xds"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(kxdsv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var (
		metricsAddr string
		probeAddr   string
		xdsAddr     string
	)
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&xdsAddr, "xds-bind-address", ":18000", "The address the xds server endpoint binds to.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		HealthProbeBindAddress: probeAddr,
		Port:                   9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	var (
		xdsCache = cache.NewSnapshotCache(
			true,
			xds.DefaultHash,
			xds.NewLogger(mgr.GetLogger()),
		)

		cacheReconciller = controllers.NewReconciler(
			mgr.GetClient(),
			xds.NewCacheRefresher(xdsCache, xds.DefautHashKey),
		)
	)

	if err := mgr.Add(xds.NewServer(xdsCache, xds.ServerConfig{BindAddr: xdsAddr})); err != nil {
		setupLog.Error(err, "unable to create the xds server")
		os.Exit(1)
	}

	// Start looking for xds services.
	if err = ctrl.NewControllerManagedBy(mgr).For(&kxdsv1alpha1.Service{}).Complete(cacheReconciller); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "kxdsv1alpha1.Service")
		os.Exit(1)
	}

	// Start looking for endpoints.
	if err = ctrl.NewControllerManagedBy(mgr).For(&corev1.Endpoints{}).Complete(cacheReconciller); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "corev1.Endpoints")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
