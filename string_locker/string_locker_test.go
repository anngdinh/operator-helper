package string_locker

import (
	"sync"
	"testing"
	"time"
)

func TestStringKeyLocker(t *testing.T) {
	locker := &StringKeyLocker{}

	// Test Lock and Unlock on a single key
	key := "test-key"
	locker.Lock(key)
	unlocked := make(chan bool)

	// Attempt to lock the same key in a separate goroutine
	go func() {
		locker.Lock(key) // Should block until unlocked
		unlocked <- true
		locker.Unlock(key)
	}()

	// Unlock the first lock after a short delay
	go func() {
		locker.Unlock(key)
	}()

	select {
	case <-unlocked:
		// Success, the second goroutine acquired the lock
	case <-time.After(1 * time.Second):
		t.Fatalf("Unlock did not allow the second goroutine to proceed.")
	}
}

func TestStringKeyLocker_ConcurrentLocks(t *testing.T) {
	locker := &StringKeyLocker{}
	keys := []string{"key1", "key2", "key3"}
	var wg sync.WaitGroup

	// Lock and unlock multiple keys concurrently
	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			locker.Lock(k)
			locker.Unlock(k)
		}(key)
	}

	wg.Wait() // Ensure all goroutines complete
}
