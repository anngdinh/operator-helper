package errs

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestHandleReconcileError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name    string
		args    args
		want    ctrl.Result
		wantErr error
	}{
		{
			name: "input err is nil",
			args: args{
				err: nil,
			},
			want:    ctrl.Result{},
			wantErr: nil,
		},
		{
			name: "input err is NeedRequeueAfter",
			args: args{
				err: NewNeedRequeueAfter("some error", 3*time.Second),
			},
			want: ctrl.Result{
				RequeueAfter: 3 * time.Second,
			},
			wantErr: nil,
		},
		{
			name: "input err is NeedRequeue",
			args: args{
				err: NewNeedRequeue("some error"),
			},
			want: ctrl.Result{
				Requeue: true,
			},
			wantErr: nil,
		},
		{
			name: "input err is other error type",
			args: args{
				err: errors.New("some error"),
			},
			want:    ctrl.Result{},
			wantErr: errors.New("some error"),
		},
	}

	logger := logrus.New().WithField("test", "TestHandleReconcileError")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleReconcileError(tt.args.err, logger)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
