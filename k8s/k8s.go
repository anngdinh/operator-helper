package k8s

import (
	"context"
	"time"

	"github.com/anngdinh/operator-helper/contexts"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// successIcon = "✅"
	// errorIcon   = "❌"
	// detectIcon  = "🔍"
	actionIcon = "🌐"
)

// EnsureObject ensures that the object exists in the cluster.
func EnsureObject(ctx context.Context, cl client.Client, obj client.Object) error {
	logger := contexts.NewContext(ctx).Log()

	// Check if the object exists
	objGet := obj.DeepCopyObject().(client.Object)
	err := cl.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, objGet)
	if err != nil && errors.IsNotFound(err) {
		logger.Infof("%s Creating object: %s/%s", actionIcon, obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
		return CreateObjectWithRetry(ctx, cl, obj)
	} else if err != nil {
		logger.Errorf("Failed to get object: %v", err)
		return err
	}

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		objGet := obj.DeepCopyObject().(client.Object)
		if err := cl.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, objGet); err != nil {
			return err
		}

		logger.Infof("%s Patching object: %s/%s/%s", actionIcon, obj.GetObjectKind().GroupVersionKind().Kind, obj.GetNamespace(), obj.GetName())
		return cl.Patch(ctx, obj, client.MergeFromWithOptions(objGet, client.MergeFromWithOptimisticLock{}))
	})
}

// create then get the object to ensure it is created
func CreateObjectWithRetry(ctx context.Context, cl client.Client, obj client.Object, opts ...client.CreateOption) error {
	logger := contexts.NewContext(ctx).Log()

	// logger.Infof("%s Creating object: %s", actionIcon, obj.GetName())
	err := cl.Create(ctx, obj, opts...)
	if err != nil {
		return err
	}

	err = errors.NewNotFound(schema.GroupResource{}, obj.GetName())
	i := 0
	// retry 3 times
	for i < 3 && err != nil && client.IgnoreNotFound(err) == nil {
		// get the object to ensure it is created
		err = cl.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, obj)
		logger.Warn("Create return nil but object not found, try to get the object again.")
		time.Sleep(250 * time.Millisecond)
		i++
	}
	if err != nil {
		return err
	}
	return nil
}

// delete then get the object to ensure it is deleted
func DeleteObjectWithRetry(ctx context.Context, cl client.Client, obj client.Object, opts ...client.DeleteOption) error {
	logger := contexts.NewContext(ctx).Log()

	// logger.Infof("%s Deleting object: %s", actionIcon, obj.GetName())
	err := cl.Delete(ctx, obj, opts...)
	if err != nil {
		return err
	}

	err = nil
	i := 0
	// retry 3 times
	for i < 3 && err != nil && client.IgnoreNotFound(err) == nil {
		// get the object to ensure it is deleted
		err = cl.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, obj)
		logger.Warn("Delete return nil but object still exists, try to get the object again.")
		time.Sleep(250 * time.Millisecond)
		i++
	}
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	return err
}
