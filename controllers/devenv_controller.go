package controllers

import (
	"context"
	"fmt"
	"time"

	appv1alpha1 "github.com/kagrabular/anareta/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"
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

	// Debug: log at method entry to see each reconcile pass
	logger.Info(">>> Entering Reconcile", "DevEnv", req.NamespacedName)

	// 1) Fetch the DevEnv instance
	var env appv1alpha1.DevEnv
	if err := r.Get(ctx, req.NamespacedName, &env); err != nil {
		if apierrors.IsNotFound(err) {
			// Resource not found; it may have been deleted after reconcile request
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 2) Handle deletion: if DeletionTimestamp is set, clean up and remove finalizer
	if !env.ObjectMeta.DeletionTimestamp.IsZero() {
		if ContainsString(env.ObjectMeta.Finalizers, finalizerName) {
			logger.Info("Cleaning up resources for DevEnv", "name", env.Name)
			// Delete the associated namespace
			if err := r.CleanupNamespace(ctx, &env); err != nil {
				logger.Error(err, "failed to delete namespace", "namespace", fmt.Sprintf("anareta-%s", env.Name))
				// Retry after a short delay on failure
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}

			// Remove finalizer
			orig := env.DeepCopy()
			env.ObjectMeta.Finalizers = RemoveString(env.ObjectMeta.Finalizers, finalizerName)
			if err := r.Patch(ctx, &env, client.MergeFrom(orig)); err != nil {
				logger.Error(err, "failed to remove finalizer from DevEnv", "name", env.Name)
				return ctrl.Result{}, err
			}
		}
		// Finalizer removed (or not present); allow Kubernetes to delete the CR
		return ctrl.Result{}, nil
	}

	// 3) Ensure finalizer is present; if not, add it and requeue
	if !ContainsString(env.ObjectMeta.Finalizers, finalizerName) {
		orig := env.DeepCopy()
		env.ObjectMeta.Finalizers = append(env.ObjectMeta.Finalizers, finalizerName)
		if err := r.Patch(ctx, &env, client.MergeFrom(orig)); err != nil {
			logger.Error(err, "unable to add finalizer to DevEnv", "name", env.Name)
			return ctrl.Result{}, err
		}
		// Requeue so that the next reconciliation sees the object with the new finalizer
		return ctrl.Result{Requeue: true}, nil
	}

	// 4) At this point, finalizer is present and we are not deleting

	// Compute namespace name
	nsName := fmt.Sprintf("anareta-%s", env.Name)

	// 4a) Ensure namespace exists
	if err := r.EnsureNamespace(ctx, nsName, &env); err != nil {
		logger.Error(err, "failed to ensure namespace", "namespace", nsName)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// 4b) Ensure Helm release (stub); if error, update status to Error
	if err := r.EnsureHelmRelease(ctx, nsName, &env); err != nil {
		logger.Error(err, "failed to ensure Helm release for DevEnv", "name", env.Name)
		// Refresh the object before updating status
		var updatedErrorEnv appv1alpha1.DevEnv
		if getErr := r.Get(ctx, req.NamespacedName, &updatedErrorEnv); getErr == nil {
			updatedErrorEnv.Status.Phase = "Error"
			updatedErrorEnv.Status.Message = err.Error()
			_ = r.Status().Update(ctx, &updatedErrorEnv)
		}
		return ctrl.Result{}, err
	}

	// 4c) Update status to Ready
	var updated appv1alpha1.DevEnv
	if err := r.Get(ctx, req.NamespacedName, &updated); err != nil {
		return ctrl.Result{}, err
	}
	if updated.Status.Phase != "Ready" {
		updated.Status.Phase = "Ready"
		updated.Status.Message = "Environment provisioned"
		updated.Status.StartedAt = metav1.Now()
		if err := r.Status().Update(ctx, &updated); err != nil {
			if apierrors.IsConflict(err) {
				// ResourceVersion conflict—requeue and try again
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
	}

	// 5) Requeue after TTL (if TTL > 0)
	if env.Spec.TTL.Duration > 0 {
		return ctrl.Result{RequeueAfter: env.Spec.TTL.Duration}, nil
	}

	return ctrl.Result{}, nil
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
		if apierrors.IsNotFound(err) {
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
	if err := r.Delete(ctx, ns); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

// EnsureHelmRelease installs or upgrades a Helm chart (stub)
func (r *DevEnvReconciler) EnsureHelmRelease(ctx context.Context, namespace string, env *appv1alpha1.DevEnv) error {
	// TODO: implement Helm SDK logic
	return nil
}
