package controllers

import (
	"context"
	"fmt"

	appv1alpha1 "github.com/Kagrabular/ANARETA/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client" // your IDE might say this is redundant, but is run in-file, leave it
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	finalizerName = "finalizer.anareta.dev"
)

// DevEnvReconciler reconciles a DevEnv object
// +kubebuilder:rbac:groups=anareta.dev,resources=devenvs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anareta.dev,resources=devenvs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=anareta.dev,resources=devenvs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;create;delete

type DevEnvReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile main reconcile loop - called by controller-runtime
func (r *DevEnvReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the DevEnv
	var env appv1alpha1.DevEnv
	if err := r.Get(ctx, req.NamespacedName, &env); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !env.ObjectMeta.DeletionTimestamp.IsZero() {
		if ContainsString(env.ObjectMeta.Finalizers, finalizerName) {
			logger.Info("Cleaning up resources for DevEnv", "name", env.Name)
			if err := r.CleanupNamespace(ctx, &env); err != nil {
				return ctrl.Result{}, err
			}
			env.ObjectMeta.Finalizers = RemoveString(env.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, &env); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Make sure finalizer
	if !ContainsString(env.ObjectMeta.Finalizers, finalizerName) {
		env.ObjectMeta.Finalizers = append(env.ObjectMeta.Finalizers, finalizerName)
		if err := r.Update(ctx, &env); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Compute namespace name
	nsName := fmt.Sprintf("anareta-%s", env.Name)

	// Ensure namespace exists
	if err := r.EnsureNamespace(ctx, nsName, nil); err != nil {
		return ctrl.Result{}, err
	}

	// Ensure Helm release (stub)
	if err := r.EnsureHelmRelease(ctx, nsName, &env); err != nil {
		env.Status.Phase = "Error"
		env.Status.Message = err.Error()
		_ = r.Status().Update(ctx, &env)
		return ctrl.Result{}, err
	}

	// Update status to Ready
	env.Status.Phase = "Ready"
	env.Status.Message = "Environment provisioned"
	env.Status.StartedAt = metav1.Now()
	if err := r.Status().Update(ctx, &env); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue after TTL
	return ctrl.Result{RequeueAfter: env.Spec.TTL.Duration}, nil
}

// SetupWithManager registers the reconciler with the manager.
func (r *DevEnvReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.DevEnv{}).
		Complete(r)
}

// EnsureNamespace creates the namespace if it doesn't exist
func (r *DevEnvReconciler) EnsureNamespace(ctx context.Context, name string, _ client.Object) error {
	ns := &corev1.Namespace{}
	key := client.ObjectKey{Name: name}
	if err := r.Get(ctx, key, ns); err != nil {
		if errors.IsNotFound(err) {
			ns = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
			return r.Create(ctx, ns)
		}
		return err
	}
	return nil
}

// CleanupNamespace deletes the namespace
func (r *DevEnvReconciler) CleanupNamespace(ctx context.Context, env *appv1alpha1.DevEnv) error {
	nsName := fmt.Sprintf("anareta-%s", env.Name)
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
	if err := r.Delete(ctx, ns); err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

// EnsureHelmRelease installs or upgrades a Helm chart
func (r *DevEnvReconciler) EnsureHelmRelease(ctx context.Context, namespace string, env *appv1alpha1.DevEnv) error {
	// TODO: implement Helm SDK logic
	return nil
}
