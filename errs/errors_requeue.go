package errs

import (
	"fmt"
	"time"
)

// NewRequeueNeeded constructs new RequeueError to
// instruct controller-runtime to requeue the processing item without been logged as error.
func NewNeedRequeue(reason string) *NeedRequeue {
	return &NeedRequeue{
		reason: reason,
	}
}

// NewRequeueNeededAfter constructs new NeedRequeueAfter to
// instruct controller-runtime to requeue the processing item after specified duration without been logged as error.
func NewNeedRequeueAfter(reason string, duration time.Duration) *NeedRequeueAfter {
	return &NeedRequeueAfter{
		reason:   reason,
		duration: duration,
	}
}

// use this when you want to requeue after a default duration
func NewNoNeedRequeue(reason string) *NoNeedRequeue {
	return &NoNeedRequeue{
		reason: reason,
	}
}

var _ error = &NeedRequeue{}

// An error to instruct controller-runtime to requeue the processing item without been logged as error.
// This should be used when a "error condition" occurrence is sort of expected and can be resolved by retry.
// e.g. a dependency haven't been fulfilled yet.
type NeedRequeue struct {
	reason string
}

func (e *NeedRequeue) Reason() string {
	return e.reason
}

func (e *NeedRequeue) Error() string {
	return fmt.Sprintf("requeue needed: %v", e.reason)
}

var _ error = &NeedRequeueAfter{}

// An error to instruct controller-runtime to requeue the processing item after specified duration without been logged as error.
// This should be used when a "error condition" occurrence is sort of expected and can be resolved by retry.
// e.g. a dependency haven't been fulfilled yet, and expected it to be fulfilled after duration.
// Note: use this with care,a simple wait might suits your use case better.
type NeedRequeueAfter struct {
	reason   string
	duration time.Duration
}

func (e *NeedRequeueAfter) Reason() string {
	return e.reason
}

func (e *NeedRequeueAfter) Duration() time.Duration {
	return e.duration
}

func (e *NeedRequeueAfter) Error() string {
	return fmt.Sprintf("requeue needed after %v: %v", e.duration, e.reason)
}

var _ error = &NoNeedRequeue{}

// An error to instruct controller-runtime to requeue the processing item after a default duration without been logged as error.
type NoNeedRequeue struct {
	reason string
}

func (e *NoNeedRequeue) Reason() string {
	return e.reason
}
func (e *NoNeedRequeue) Error() string {
	return fmt.Sprintf("no need to requeue: %v", e.reason)
}
