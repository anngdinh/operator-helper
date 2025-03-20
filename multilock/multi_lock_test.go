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
