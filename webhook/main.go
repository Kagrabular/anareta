package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1alpha1 "github.com/Kagrabular/ANARETA/api/v1alpha1"
	"github.com/Kagrabular/ANARETA/webhook"
)

var (
	scheme = func() *runtime.Scheme {
		s := runtime.NewScheme()
		utilruntime.Must(clientgoscheme.AddToScheme(s))
		utilruntime.Must(appv1alpha1.AddToScheme(s))
		return s
	}()
)

func main() {
	var (
		addr string
		ns   string
	)
	flag.StringVar(&addr, "listen-addr", ":8080", "address to listen on")
	flag.StringVar(&ns, "namespace", "default", "namespace for DevEnv resources")
	flag.Parse()

	cfg := ctrl.GetConfigOrDie()
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}

	http.HandleFunc("/webhook", handlers.MakeWebhookHandler(k8sClient, ns))
	srv := &http.Server{
		Addr:         addr,
		Handler:      nil,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Printf("Starting webhook server on %s\n", addr)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
