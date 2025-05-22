package integration_test

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	anareta "github.com/Kagrabular/anareta/api/v1alpha1"
	controllers "github.com/Kagrabular/anareta/controllers" // ide will call this redundant, but one mans redundancy is another mans is a package used in file. don't worry about this.
)

func TestReconcileCreatesNamespace(t *testing.T) {
	// guarantee manager controllers tore down before invoke testEnv.Stop(), goroutines still had open watches and handlers bound to api-server, call to kube-api server hangs up waiting for connections to close.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{"../../config/crd/bases"},
	}
	cfg, err := testEnv.Start()
	require.NoError(t, err)
	defer func() {
		// first stop the manager
		cancel() // signaling mgr.Start to exit
		// then shut down the control plane
		if err := testEnv.Stop(); err != nil {
			t.Fatalf("Failed to stop envtest: %v", err)
		}
	}()

	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	require.NoError(t, anareta.AddToScheme(scheme))

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	require.NoError(t, err)

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme})
	require.NoError(t, err)

	reconciler := &controllers.DevEnvReconciler{
		Client: k8sClient,
		Scheme: mgr.GetScheme(),
	}
	require.NoError(t, reconciler.SetupWithManager(mgr))

	go func() {
		_ = mgr.Start(ctx)
	}()

	dev := &anareta.DevEnv{
		ObjectMeta: metav1.ObjectMeta{Name: "integ-test", Namespace: "default"},
		Spec: anareta.DevEnvSpec{
			RepoURL: "https://github.com/Kagrabular/myapp.git",
			Branch:  "main",
			TTL:     metav1.Duration{Duration: 1 * time.Minute},
		},
	}
	require.NoError(t, k8sClient.Create(ctx, dev))

	ns := &corev1.Namespace{}
	err = wait.PollImmediate(100*time.Millisecond, 5*time.Second, func() (bool, error) {
		getErr := k8sClient.Get(ctx, client.ObjectKey{Name: "anareta-integ-test"}, ns)
		if getErr != nil {
			return false, nil
		}
		return true, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "anareta-integ-test", ns.Name)
}
