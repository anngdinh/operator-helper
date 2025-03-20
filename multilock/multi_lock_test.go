package multilock

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMultiLockGoroutineLockSameKey(t *testing.T) {
	database := []string{}
	locker := NewMultipleLock()
	addData := func(name string) {
		locker.Lock("key")
		defer locker.Unlock("key")
		time.Sleep(100 * time.Millisecond)
		database = append(database, name)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			addData(fmt.Sprintf("goroutine-%d", i))
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println(database)
	if len(database) != 10 {
		t.Fatalf("Expected 10 items in database, got %d", len(database))
	}
}

func TestStringKeyLocker(t *testing.T) {
	locker := NewMultipleLock()

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
	locker := NewMultipleLock()
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

// func TestStringKeyLocker_UnlockManyTimes(t *testing.T) {
// 	locker := NewMultipleLock()

// 	locker.Lock("key")
// 	locker.Unlock("key")
// 	locker.Unlock("key") // Should not panic
// }
