package string_locker

import (
	"sync"
)

// StringKeyLocker manages locks based on string keys.
type StringKeyLocker struct {
	locks sync.Map // Map of string keys to sync.Mutex
}

// Lock locks the mutex for a given key.
func (l *StringKeyLocker) Lock(key string) {
	val, _ := l.locks.LoadOrStore(key, &sync.Mutex{})
	mutex := val.(*sync.Mutex)
	mutex.Lock()
}

// Unlock unlocks the mutex for a given key.
func (l *StringKeyLocker) Unlock(key string) {
	val, ok := l.locks.Load(key)
	if !ok {
		// panic(fmt.Sprintf("attempt to unlock a non-existent key: %s", key))
		return
	}
	mutex := val.(*sync.Mutex)
	mutex.Unlock()
	l.locks.Delete(key)
}
