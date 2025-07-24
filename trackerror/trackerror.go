package trackerror

import "sync"

type TrackError interface {
	Load(key string) bool
	Store(key string)
	Delete(key string)
	Count() int
}

var _ TrackError = &trackError{}

type trackError struct {
	trackMap map[string]bool
	count    int
	sync.RWMutex
}

func NewTrackError() TrackError {
	return &trackError{
		trackMap: make(map[string]bool),
		count:    0,
	}
}

func (t *trackError) Load(key string) bool {
	t.RWMutex.RLock()
	defer t.RWMutex.RUnlock()
	if isError, ok := t.trackMap[key]; ok {
		return isError
	}
	return false
}

func (t *trackError) Store(key string) {
	t.RWMutex.Lock()
	defer t.RWMutex.Unlock()
	if _, ok := t.trackMap[key]; !ok {
		t.trackMap[key] = true
		t.count++
	}
}

func (t *trackError) Delete(key string) {
	t.RWMutex.Lock()
	defer t.RWMutex.Unlock()
	if _, ok := t.trackMap[key]; ok {
		delete(t.trackMap, key)
		t.count--
	}
}

func (t *trackError) Count() int {
	t.RWMutex.RLock()
	defer t.RWMutex.RUnlock()
	return t.count
}
