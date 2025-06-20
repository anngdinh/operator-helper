package event_classification

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MockObject is a mock implementation of client.Object for testing purposes
type MockObject struct {
	client.Object
	name              string
	deletionTimestamp *v1.Time
}

func (m *MockObject) GetDeletionTimestamp() *v1.Time {
	return m.deletionTimestamp
}

func (m *MockObject) GetName() string {
	return m.name
}

func (m *MockObject) DeepCopyObject() runtime.Object {
	return m
}

// Test cases
func TestEventClassification_Classify(t *testing.T) {
	// Helper functions
	isValid := func(obj client.Object) bool {
		return obj != nil && obj.GetName() != ""
	}

	getResourceByKey := func(key string) (client.Object, bool) {
		if key == "valid-resource" {
			return &MockObject{name: "valid-resource"}, true
		}
		if key == "deleted-resource" {
			now := v1.Now()
			return &MockObject{name: "deleted-resource", deletionTimestamp: &now}, true
		}
		return nil, false
	}

	// Create an EventClassification instance
	ec := NewEventClassification(getResourceByKey, isValid)

	ctx := context.TODO()
	// Test CreateEvent
	t.Run("CreateEvent", func(t *testing.T) {
		event := ec.Classify(ctx, "valid-resource")
		assert.NotNil(t, event)
		assert.Equal(t, CreateEvent, event.Type)
		assert.Equal(t, "valid-resource", event.Obj.GetName())
	})

	// Test DeleteEvent (when object is deleted in API, but exists in cache)
	t.Run("DeleteEventFromCache", func(t *testing.T) {
		ec.cache["deleted-resource"] = &MockObject{name: "deleted-resource"}
		event := ec.Classify(ctx, "deleted-resource")
		assert.NotNil(t, event)
		assert.Equal(t, DeleteEvent, event.Type)
		assert.Equal(t, "deleted-resource", event.Obj.GetName())
	})

	// Test DeleteEvent (when resource has deletion timestamp)
	t.Run("DeleteEventWithDeletionTimestamp", func(t *testing.T) {
		event := ec.Classify(ctx, "deleted-resource")
		assert.NotNil(t, event)
		assert.Equal(t, DeleteEvent, event.Type)
		assert.Equal(t, "deleted-resource", event.Obj.GetName())
	})

	// Test SyncEvent
	t.Run("SyncEvent", func(t *testing.T) {
		ec.cache["valid-resource"] = &MockObject{name: "valid-resource"}
		event := ec.Classify(ctx, "valid-resource")
		assert.NotNil(t, event)
		assert.Equal(t, SyncEvent, event.Type)
		assert.Equal(t, "valid-resource", event.Obj.GetName())
		assert.Equal(t, "valid-resource", event.OldObj.GetName())
	})

	// Test invalid case (both objects invalid)
	t.Run("InvalidEvent", func(t *testing.T) {
		invalidGetResourceByKey := func(key string) (client.Object, bool) {
			return &MockObject{name: ""}, true
		}
		invalidEC := NewEventClassification(invalidGetResourceByKey, isValid)
		event := invalidEC.Classify(ctx, "invalid-resource")
		assert.Nil(t, event)
	})
}

func TestEventClassification_Stress(t *testing.T) {
	// Helper functions
	isValid := func(obj client.Object) bool {
		return obj != nil && obj.GetName() != ""
	}

	getResourceByKey := func(key string) (client.Object, bool) {
		if key == "valid-resource" {
			return &MockObject{name: "valid-resource"}, true
		}
		if key == "deleted-resource" {
			now := v1.Now()
			return &MockObject{name: "deleted-resource", deletionTimestamp: &now}, true
		}
		return nil, false
	}

	// Create an EventClassification instance
	ec := NewEventClassification(getResourceByKey, isValid)
	ctx := context.TODO()

	// Number of concurrent operations
	numGoroutines := 100
	numOperations := 1000

	// Channel to collect results
	results := make(chan *Event, numGoroutines*numOperations)
	errors := make(chan error, numGoroutines*numOperations)

	// Start timing
	start := time.Now()

	// Launch goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				// Alternate between different types of resources
				resourceKey := "valid-resource"
				if j%3 == 0 {
					resourceKey = "deleted-resource"
				}

				event := ec.Classify(ctx, resourceKey)
				if event == nil {
					errors <- fmt.Errorf("goroutine %d: got nil event for resource %s", id, resourceKey)
					continue
				}

				// Verify event properties
				if event.Obj == nil {
					errors <- fmt.Errorf("goroutine %d: event.Obj is nil", id)
					continue
				}

				if event.Obj.GetName() != resourceKey {
					errors <- fmt.Errorf("goroutine %d: expected name %s, got %s", id, resourceKey, event.Obj.GetName())
					continue
				}

				results <- event
			}
		}(i)
	}

	// Collect results
	var eventCounts = make(map[EventType]int)
	for i := 0; i < numGoroutines*numOperations; i++ {
		select {
		case event := <-results:
			eventCounts[event.Type]++
		case err := <-errors:
			t.Errorf("Error during stress test: %v", err)
		}
	}

	// Calculate duration
	duration := time.Since(start)

	// Verify results
	t.Logf("Stress test completed in %v", duration)
	t.Logf("Total events processed: %d", numGoroutines*numOperations)
	t.Logf("Event type distribution: %+v", eventCounts)

	// Verify we have a reasonable distribution of event types
	assert.Greater(t, eventCounts[CreateEvent], 0, "Should have some create events")
	assert.Greater(t, eventCounts[DeleteEvent], 0, "Should have some delete events")
	assert.Greater(t, eventCounts[SyncEvent], 0, "Should have some sync events")

	// Verify total count
	totalEvents := 0
	for _, count := range eventCounts {
		totalEvents += count
	}
	assert.Equal(t, numGoroutines*numOperations, totalEvents, "Total events should match expected count")
}

func TestEventClassification_EmptyKeyGuards(t *testing.T) {
	isValid := func(obj client.Object) bool {
		return obj != nil && obj.GetName() != ""
	}
	getResourceByKey := func(key string) (client.Object, bool) {
		return &MockObject{name: "some-resource"}, true
	}
	// Create an EventClassification instance
	ec := NewEventClassification(getResourceByKey, isValid)
	ctx := context.TODO()

	t.Run("Classify with empty key returns nil", func(t *testing.T) {
		event := ec.Classify(ctx, "")
		assert.Nil(t, event)
	})

	t.Run("writeCache with empty key does not panic or write", func(t *testing.T) {
		ec.writeCache("", &MockObject{name: "should-not-write"})
		_, exists := ec.cache[""]
		assert.False(t, exists)
	})

	t.Run("readCache with empty key returns nil, false", func(t *testing.T) {
		obj, ok := ec.readCache("")
		assert.Nil(t, obj)
		assert.False(t, ok)
	})

	t.Run("deleteCache with empty key does not panic", func(t *testing.T) {
		// Should not panic or affect the map
		ec.cache["foo"] = &MockObject{name: "foo"}
		ec.deleteCache("")
		_, exists := ec.cache["foo"]
		assert.True(t, exists)
	})
}

func TestEventClassification_Classify_AllBranches(t *testing.T) {
	ctx := context.TODO()

	// 1. Both cache and get are missing: !okCache && !okGet
	{
		isValid := func(obj client.Object) bool { return false }
		getResourceByKey := func(key string) (client.Object, bool) { return nil, false }
		ec := NewEventClassification(getResourceByKey, isValid)
		event := ec.Classify(ctx, "missing-resource")
		assert.Nil(t, event)
	}

	// 2. Both cache and get exist, but both are invalid: okCache && okGet && !objGetValid && !objCacheValid
	{
		isValid := func(obj client.Object) bool { return false }
		getResourceByKey := func(key string) (client.Object, bool) { return &MockObject{name: "invalid"}, true }
		ec := NewEventClassification(getResourceByKey, isValid)
		ec.cache["invalid-resource"] = &MockObject{name: "invalid"}
		event := ec.Classify(ctx, "invalid-resource")
		assert.Nil(t, event)
	}

	// 3. okCache && okGet, objGetValid is false, objCacheValid is true: should trigger DeleteEvent
	{
		isValid := func(obj client.Object) bool {
			if obj == nil {
				return false
			}
			return obj.GetName() == "valid" // only cache is valid
		}
		getResourceByKey := func(key string) (client.Object, bool) { return &MockObject{name: "invalid"}, true }
		ec := NewEventClassification(getResourceByKey, isValid)
		ec.cache["resource"] = &MockObject{name: "valid"}
		event := ec.Classify(ctx, "resource")
		assert.NotNil(t, event)
		assert.Equal(t, DeleteEvent, event.Type)
		assert.Equal(t, "valid", event.Obj.GetName())
	}

	// 4. okCache && okGet, objGetValid is true, objCacheValid is false: should trigger CreateEvent
	{
		isValid := func(obj client.Object) bool {
			if obj == nil {
				return false
			}
			return obj.GetName() == "valid" // only get is valid
		}
		getResourceByKey := func(key string) (client.Object, bool) { return &MockObject{name: "valid"}, true }
		ec := NewEventClassification(getResourceByKey, isValid)
		ec.cache["resource"] = &MockObject{name: "invalid"}
		event := ec.Classify(ctx, "resource")
		assert.NotNil(t, event)
		assert.Equal(t, CreateEvent, event.Type)
		assert.Equal(t, "valid", event.Obj.GetName())
	}
}
