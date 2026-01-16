package workerpool

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestPool_AddTaskAndWait(t *testing.T) {
	var counter int32
	pool := NewPool(3, BackoffConfig{
		InitialDelay: 10 * time.Millisecond,
		MaxRetries:   3,
		Factor:       2.0,
	})
	pool.Start()

	taskCount := 10
	for i := 0; i < taskCount; i++ {
		pool.AddTask(NewTask(func() error {
			atomic.AddInt32(&counter, 1)
			return nil
		}))
	}

	pool.Wait()
	pool.Close()

	if counter != int32(taskCount) {
		t.Errorf("expected counter to be %d, got %d", taskCount, counter)
	}
}

func TestPool_ConcurrentAddTask(t *testing.T) {
	var counter int32
	pool := NewPool(5, BackoffConfig{
		InitialDelay: 10 * time.Millisecond,
		MaxRetries:   3,
		Factor:       2.0,
	})
	pool.Start()

	taskCount := 100
	done := make(chan struct{})

	go func() {
		for i := 0; i < taskCount; i++ {
			pool.AddTask(NewTask(func() error {
				time.Sleep(1 * time.Millisecond)
				atomic.AddInt32(&counter, 1)
				return nil
			}))
		}
		close(done)
	}()

	<-done
	pool.Wait()
	pool.Close()

	if counter != int32(taskCount) {
		t.Errorf("expected counter to be %d, got %d", taskCount, counter)
	}
}

func TestPool_Close(t *testing.T) {
	pool := NewPool(2, BackoffConfig{
		InitialDelay: 10 * time.Millisecond,
		MaxRetries:   3,
		Factor:       2.0,
	})
	pool.Start()
	pool.Close()
	// Closing twice should not panic
	pool.Close()
}

func TestPool_Run_AllTasksExecuted(t *testing.T) {
	taskCount := 10
	var executed int32
	tasks := make([]*Task, 0, taskCount)

	for i := 0; i < taskCount; i++ {
		tasks = append(tasks, NewTask(func() error {
			atomic.AddInt32(&executed, 1)
			return nil
		}))
	}

	backoff := BackoffConfig{
		InitialDelay: 10 * time.Millisecond,
		MaxRetries:   3,
		Factor:       2.0,
	}
	pool := NewPool(3, backoff)
	pool.Start()
	for _, task := range tasks {
		pool.AddTask(task)
	}
	pool.Close()
	pool.Wait()

	if int(executed) != taskCount {
		t.Errorf("expected %d tasks executed, got %d", taskCount, executed)
	}
}

func TestPool_Run_ErrorsCaptured(t *testing.T) {
	backoff := BackoffConfig{
		InitialDelay: 10 * time.Millisecond,
		MaxRetries:   3,
		Factor:       2.0,
	}
	tasks := []*Task{
		NewTask(func() error { return nil }),
		NewTask(func() error { return errors.New("foo error") }),
	}
	pool := NewPool(2, backoff)
	pool.Start()
	for _, task := range tasks {
		pool.AddTask(task)
	}
	pool.Close()
	pool.Wait()

	if tasks[0].Err != nil {
		t.Errorf("expected no error for first task, got %v", tasks[0].Err)
	}
	if tasks[1].Err == nil {
		t.Errorf("expected error for second task, got nil")
	}
}

func TestTask_RetryOnErrorWithBackoff(t *testing.T) {
	var attempts int32
	backoff := BackoffConfig{
		InitialDelay: 10 * time.Millisecond,
		MaxRetries:   3,
		Factor:       2.0,
	}
	task := NewTask(func() error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("foo error")
	})

	pool := NewPool(1, backoff)
	pool.Start()
	pool.AddTask(task)
	pool.Close()
	pool.Wait()

	if attempts != 4 {
		t.Errorf("expected 4 attempts, got %d", attempts)
	}
	if task.Err == nil {
		t.Errorf("expected error after retries, got nil")
	}
}
