package trackerror

import (
	"sync"
	"sync/atomic"
)

type SyncMapTrackError struct {
	trackMap sync.Map
	count    int64
}

func NewSyncMapTrackError() TrackError {
	return &SyncMapTrackError{}
}

func (t *SyncMapTrackError) Load(key string) bool {
	_, ok := t.trackMap.Load(key)
	return ok
}

func (t *SyncMapTrackError) Store(key string) {
	_, loaded := t.trackMap.LoadOrStore(key, true)
	if !loaded {
		atomic.AddInt64(&t.count, 1)
	}
}

func (t *SyncMapTrackError) Delete(key string) {
	_, deleted := t.trackMap.LoadAndDelete(key)
	if deleted {
		atomic.AddInt64(&t.count, -1)
	}
}

func (t *SyncMapTrackError) Count() int {
	return int(atomic.LoadInt64(&t.count))
}