package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestEnsureObject_Create(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()

	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-cm",
			Namespace: "default",
		},
		Data: map[string]string{"key": "value"},
	}

	err := EnsureObject(ctx, cl, cm)
	require.NoError(t, err)

	var fetched corev1.ConfigMap
	err = cl.Get(ctx, types.NamespacedName{Name: "test-cm", Namespace: "default"}, &fetched)
	require.NoError(t, err)
	assert.Equal(t, "value", fetched.Data["key"])
}

func TestCreateObjectWithRetry(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()

	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "retry-cm",
			Namespace: "default",
		},
	}

	err := CreateObjectWithRetry(ctx, cl, cm)
	require.NoError(t, err)

	var fetched corev1.ConfigMap
	err = cl.Get(ctx, types.NamespacedName{Name: "retry-cm", Namespace: "default"}, &fetched)
	require.NoError(t, err)
}

func TestDeleteObjectWithRetry(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "delete-cm",
			Namespace: "default",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm).Build()
	ctx := context.Background()

	err := DeleteObjectWithRetry(ctx, cl, cm)
	require.NoError(t, err)

	var fetched corev1.ConfigMap
	err = cl.Get(ctx, types.NamespacedName{Name: "delete-cm", Namespace: "default"}, &fetched)
	assert.Error(t, err)
	assert.True(t, client.IgnoreNotFound(err) == nil)
}

func TestCreateObjectWithRetry_GenerateName(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()

	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "gen-cm-",
			Namespace:    "default",
		},
		Data: map[string]string{"generated": "true"},
	}

	err := CreateObjectWithRetry(ctx, cl, cm)
	require.NoError(t, err)

	assert.NotEmpty(t, cm.Name, "Name should be populated after creation")

	var fetched corev1.ConfigMap
	err = cl.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: "default"}, &fetched)
	require.NoError(t, err)
	assert.Equal(t, "true", fetched.Data["generated"])
}
