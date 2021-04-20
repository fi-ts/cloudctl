package cmd

import (
	"reflect"
	"testing"

	"github.com/fi-ts/cloud-go/api/models"
)

func Test_parsePipe(t *testing.T) {
	type args struct {
		unparsed string
	}
	tests := []struct {
		name    string
		args    args
		want    *models.V1PipeSpec
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				unparsed: "nginx:8080:cluster-int-nginx:8082",
			},
			want: &models.V1PipeSpec{
				Name:   ptr("nginx"),
				Port:   i64Ptr(8080),
				Remote: ptr("cluster-int-nginx:8082"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePipe(tt.args.unparsed)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePipe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func i64Ptr(i int) *int64 {
	i64 := int64(i)
	return &i64
}
