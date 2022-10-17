package log

import (
	"context"
	"reflect"
	"testing"

	"go.uber.org/zap"
)

func Test_esimZap_getArgs(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	ctx := context.Background()

	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{"解析路径", args{ctx}, []interface{}{"caller", "testing/testing.go:991"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ez := NewEsimZap()
			if got := ez.getArgs(tt.args.ctx, zap.InfoLevel); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("esimZap.getArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
