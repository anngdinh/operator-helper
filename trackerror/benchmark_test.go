package trackerror

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

func BenchmarkTrackError_Store(b *testing.B) {
	tracker := NewTrackError()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}
}

func BenchmarkSyncMapTrackError_Store(b *testing.B) {
	tracker := NewSyncMapTrackError()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}
}

func BenchmarkTrackError_Load(b *testing.B) {
	tracker := NewTrackError()

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		tracker.Load(key)
	}
}

func BenchmarkSyncMapTrackError_Load(b *testing.B) {
	tracker := NewSyncMapTrackError()

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		tracker.Load(key)
	}
}

func BenchmarkTrackError_Delete(b *testing.B) {
	tracker := NewTrackError()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Delete(key)
	}
}

func BenchmarkSyncMapTrackError_Delete(b *testing.B) {
	tracker := NewSyncMapTrackError()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Delete(key)
	}
}

func BenchmarkTrackError_Count(b *testing.B) {
	tracker := NewTrackError()

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tracker.Count()
	}
}

func BenchmarkSyncMapTrackError_Count(b *testing.B) {
	tracker := NewSyncMapTrackError()

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tracker.Count()
	}
}

func BenchmarkTrackError_ConcurrentStore(b *testing.B) {
	tracker := NewTrackError()
	numCPU := runtime.NumCPU()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d-%d", i%numCPU, i)
			tracker.Store(key)
			i++
		}
	})
}

func BenchmarkSyncMapTrackError_ConcurrentStore(b *testing.B) {
	tracker := NewSyncMapTrackError()
	numCPU := runtime.NumCPU()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d-%d", i%numCPU, i)
			tracker.Store(key)
			i++
		}
	})
}

func BenchmarkTrackError_ConcurrentLoad(b *testing.B) {
	tracker := NewTrackError()

	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%10000)
			tracker.Load(key)
			i++
		}
	})
}

func BenchmarkSyncMapTrackError_ConcurrentLoad(b *testing.B) {
	tracker := NewSyncMapTrackError()

	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%10000)
			tracker.Load(key)
			i++
		}
	})
}

func BenchmarkTrackError_ConcurrentMixed(b *testing.B) {
	tracker := NewTrackError()

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%1000)
			switch i % 4 {
			case 0:
				tracker.Store(key)
			case 1:
				tracker.Load(key)
			case 2:
				tracker.Count()
			case 3:
				if i%10 == 0 {
					tracker.Delete(key)
				}
			}
			i++
		}
	})
}

func BenchmarkSyncMapTrackError_ConcurrentMixed(b *testing.B) {
	tracker := NewSyncMapTrackError()

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		tracker.Store(key)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%1000)
			switch i % 4 {
			case 0:
				tracker.Store(key)
			case 1:
				tracker.Load(key)
			case 2:
				tracker.Count()
			case 3:
				if i%10 == 0 {
					tracker.Delete(key)
				}
			}
			i++
		}
	})
}

func BenchmarkTrackError_HighContention(b *testing.B) {
	tracker := NewTrackError()
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU() * 2

	b.ResetTimer()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < b.N/numWorkers; j++ {
				key := fmt.Sprintf("shared_key_%d", j%100)
				tracker.Store(key)
				tracker.Load(key)
				if j%5 == 0 {
					tracker.Delete(key)
				}
				tracker.Count()
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkSyncMapTrackError_HighContention(b *testing.B) {
	tracker := NewSyncMapTrackError()
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU() * 2

	b.ResetTimer()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < b.N/numWorkers; j++ {
				key := fmt.Sprintf("shared_key_%d", j%100)
				tracker.Store(key)
				tracker.Load(key)
				if j%5 == 0 {
					tracker.Delete(key)
				}
				tracker.Count()
			}
		}(i)
	}

	wg.Wait()
}
