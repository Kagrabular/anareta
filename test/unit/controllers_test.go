// test/unit/controllers_test.go
package controllers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	anareta "github.com/Kagrabular/ANARETA/api/v1alpha1"
	controllers "github.com/Kagrabular/ANARETA/controllers"
)

func TestContainsRemoveString(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.True(t, controllers.ContainsString(slice, "b"))
	assert.False(t, controllers.ContainsString(slice, "x"))

	removed := controllers.RemoveString(slice, "b")
	assert.Len(t, removed, 2)
	assert.NotContains(t, removed, "b")
}

func TestEnsureAndCleanupNamespace(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = anareta.AddToScheme(scheme)

	// Fake client with no existing namespace
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	reconciler := &controllers.DevEnvReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	ctx := context.Background()
	nsName := "anareta-test-env"
	// Create a dummy DevEnv for owner reference
	dev := &anareta.DevEnv{
		ObjectMeta: metav1.ObjectMeta{Name: "test-env", Namespace: "default"},
	}

	// Test ensureNamespace
	err := reconciler.EnsureNamespace(ctx, nsName, dev)
	assert.NoError(t, err)

	// The namespace should now exist, hopefully
	ns := &corev1.Namespace{}
	err = fakeClient.Get(ctx, client.ObjectKey{Name: nsName}, ns)
	assert.NoError(t, err)
	assert.Equal(t, nsName, ns.Name)

	// Test cleanupNamespace
	env := anareta.DevEnv{ObjectMeta: metav1.ObjectMeta{Name: "test-env"}}
	// Populate a namespace to delete
	ns = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
	_ = fakeClient.Create(ctx, ns)

	err = reconciler.CleanupNamespace(ctx, &env)
	assert.NoError(t, err)
	// Cleanup should remove namespace
	err = fakeClient.Get(ctx, client.ObjectKey{Name: nsName}, &corev1.Namespace{})
	assert.Error(t, err)
}
