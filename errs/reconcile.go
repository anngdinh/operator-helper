package errs

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

// HandleReconcileError will handle errors from reconcile handlers, which respects runtime errors.
func HandleReconcileError(err error, log *logrus.Entry) (ctrl.Result, error) {
	if err == nil {
		return ctrl.Result{}, nil
	}

	var requeueNeededAfter *NeedRequeueAfter
	if errors.As(err, &requeueNeededAfter) {
		log.Info("requeue after duration: ", requeueNeededAfter.Duration(), ", reason: ", requeueNeededAfter.Reason())
		return ctrl.Result{RequeueAfter: requeueNeededAfter.Duration()}, nil
	}

	var requeueNeeded *NeedRequeue
	if errors.As(err, &requeueNeeded) {
		log.Info("requeue immediately reason: ", requeueNeeded.Reason())
		return ctrl.Result{Requeue: true}, nil
	}

	var noNeedRequeue *NoNeedRequeue
	if errors.As(err, &noNeedRequeue) {
		log.Info("no need to requeue, reason: ", noNeedRequeue.Reason())
		return ctrl.Result{}, nil
	}

	log.Infof("requeue after 5 seconds + exponential back-off, reason: %v", err)
	time.Sleep(5 * time.Second)
	return ctrl.Result{}, err
}

/*
ctrl.Result{}
	- Requeue bool
	- RequeueAfter time.Duration
error
	- nil
	- not nil
*/
