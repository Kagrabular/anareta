package main

import (
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	anaretaapi "github.com/Kagrabular/anareta/api/v1alpha1"
	"github.com/Kagrabular/anareta/controllers"
)

var (
	scheme = runtime.NewScheme()
	// You can keep syncPeriod defined here for later, but you won't reference it in ctrl.Options.
	syncPeriod = time.Minute
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = anaretaapi.AddToScheme(scheme)
}

func main() {
	// Set up a logger for the manager
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Create the manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,

		// Serve Prometheus metrics on :8080
		Metrics: metricsserver.Options{
			BindAddress: ":8080",
		},

		// Serve health/readiness probes on :9443
		HealthProbeBindAddress: ":9443",

		// Note: no SyncPeriod here

		// Leader election (if you ever enable it)
		LeaderElection:   false,
		LeaderElectionID: "anareta-operator-lock",
	})
	if err != nil {
		os.Exit(1)
	}

	// Register the DevEnv controller with the manager
	if err := (&controllers.DevEnvReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		os.Exit(1)
	}

	// Add health check endpoint (/healthz)
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		os.Exit(1)
	}

	// Add readiness check endpoint (/readyz)
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		os.Exit(1)
	}

	// Start the manager (this will block until a termination signal is received)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		os.Exit(1)
	}
}
