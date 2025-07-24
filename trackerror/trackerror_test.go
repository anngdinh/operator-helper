package trackerror

import (
	"sync"
	"testing"
)

func TestNewTrackError(t *testing.T) {
	tracker := NewTrackError()
	if tracker == nil {
		t.Fatal("NewTrackError() returned nil")
	}

	if tracker.Count() != 0 {
		t.Errorf("NewTrackError() count = %d, want 0", tracker.Count())
	}
}

func TestTrackError_Store(t *testing.T) {
	tracker := NewTrackError()

	tracker.Store("key1")
	if tracker.Count() != 1 {
		t.Errorf("After Store, count = %d, want 1", tracker.Count())
	}

	tracker.Store("key2")
	if tracker.Count() != 2 {
		t.Errorf("After Store, count = %d, want 2", tracker.Count())
	}

	tracker.Store("key1")
	if tracker.Count() != 2 {
		t.Errorf("After Store duplicate key, count = %d, want 2", tracker.Count())
	}
}

func TestTrackError_Load(t *testing.T) {
	tracker := NewTrackError()

	if tracker.Load("nonexistent") {
		t.Error("Load() for nonexistent key = true, want false")
	}

	tracker.Store("key1")
	if !tracker.Load("key1") {
		t.Error("Load() for existing key = false, want true")
	}

	if tracker.Load("key2") {
		t.Error("Load() for different nonexistent key = true, want false")
	}
}

func TestTrackError_Delete(t *testing.T) {
	tracker := NewTrackError()

	tracker.Delete("nonexistent")
	if tracker.Count() != 0 {
		t.Errorf("After Delete nonexistent key, count = %d, want 0", tracker.Count())
	}

	tracker.Store("key1")
	tracker.Store("key2")
	if tracker.Count() != 2 {
		t.Errorf("After Store, count = %d, want 2", tracker.Count())
	}

	tracker.Delete("key1")
	if tracker.Count() != 1 {
		t.Errorf("After Delete, count = %d, want 1", tracker.Count())
	}

	if tracker.Load("key1") {
		t.Error("Load() for deleted key = true, want false")
	}

	if !tracker.Load("key2") {
		t.Error("Load() for remaining key = false, want true")
	}

	tracker.Delete("key1")
	if tracker.Count() != 1 {
		t.Errorf("After Delete already deleted key, count = %d, want 1", tracker.Count())
	}
}

func TestTrackError_Count(t *testing.T) {
	tracker := NewTrackError()

	if tracker.Count() != 0 {
		t.Errorf("Initial count = %d, want 0", tracker.Count())
	}

	tracker.Store("key1")
	tracker.Store("key2")
	tracker.Store("key3")
	if tracker.Count() != 3 {
		t.Errorf("After 3 stores, count = %d, want 3", tracker.Count())
	}

	tracker.Delete("key2")
	if tracker.Count() != 2 {
		t.Errorf("After delete, count = %d, want 2", tracker.Count())
	}

	tracker.Delete("key1")
	tracker.Delete("key3")
	if tracker.Count() != 0 {
		t.Errorf("After deleting all, count = %d, want 0", tracker.Count())
	}
}

func TestTrackError_ConcurrentAccess(t *testing.T) {
	tracker := NewTrackError()
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "key" + string(rune(id))

			tracker.Store(key)
			tracker.Load(key)
			tracker.Count()

			if id%2 == 0 {
				tracker.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	finalCount := tracker.Count()
	if finalCount < 0 || finalCount > numGoroutines {
		t.Errorf("Final count %d is out of expected range [0, %d]", finalCount, numGoroutines)
	}
}

func TestTrackError_ConcurrentStoreLoad(t *testing.T) {
	tracker := NewTrackError()
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			tracker.Store("concurrent_key")
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			tracker.Load("concurrent_key")
		}
	}()

	wg.Wait()

	if tracker.Count() != 1 {
		t.Errorf("After concurrent operations, count = %d, want 1", tracker.Count())
	}

	if !tracker.Load("concurrent_key") {
		t.Error("Load() after concurrent operations = false, want true")
	}
}

func TestTrackError_ConcurrentStoreDelete(t *testing.T) {
	tracker := NewTrackError()
	var wg sync.WaitGroup

	tracker.Store("test_key")

	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			tracker.Store("test_key")
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			if i == 50 {
				tracker.Delete("test_key")
			}
		}
	}()

	wg.Wait()

	count := tracker.Count()
	if count < 0 || count > 1 {
		t.Errorf("Final count %d is out of expected range [0, 1]", count)
	}
}

func TestTrackError_Interface(t *testing.T) {
	var _ TrackError = NewTrackError()
}

func TestTrackError_EmptyKey(t *testing.T) {
	tracker := NewTrackError()

	tracker.Store("")
	if tracker.Count() != 1 {
		t.Errorf("After storing empty key, count = %d, want 1", tracker.Count())
	}

	if !tracker.Load("") {
		t.Error("Load() for empty key = false, want true")
	}

	tracker.Delete("")
	if tracker.Count() != 0 {
		t.Errorf("After deleting empty key, count = %d, want 0", tracker.Count())
	}
}

func TestTrackError_MultipleOperations(t *testing.T) {
	tracker := NewTrackError()

	keys := []string{"a", "b", "c", "d", "e"}

	for _, key := range keys {
		tracker.Store(key)
	}

	if tracker.Count() != len(keys) {
		t.Errorf("After storing %d keys, count = %d, want %d", len(keys), tracker.Count(), len(keys))
	}

	for _, key := range keys {
		if !tracker.Load(key) {
			t.Errorf("Load() for key %s = false, want true", key)
		}
	}

	tracker.Delete("b")
	tracker.Delete("d")

	if tracker.Count() != len(keys)-2 {
		t.Errorf("After deleting 2 keys, count = %d, want %d", tracker.Count(), len(keys)-2)
	}

	if tracker.Load("b") {
		t.Error("Load() for deleted key 'b' = true, want false")
	}

	if tracker.Load("d") {
		t.Error("Load() for deleted key 'd' = true, want false")
	}

	if !tracker.Load("a") || !tracker.Load("c") || !tracker.Load("e") {
		t.Error("Load() for remaining keys should return true")
	}
}
