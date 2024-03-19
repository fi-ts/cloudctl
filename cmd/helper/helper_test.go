package helper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandHomeDir(t *testing.T) {
	homedir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "no home dir",
			path:    "/my/absolute/path",
			want:    "/my/absolute/path",
			wantErr: false,
		},
		{
			name:    "with home dir",
			path:    "~/file",
			want:    homedir + "/file",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandHomeDir(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandHomeDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExpandHomeDir() = %v, want %v", got, tt.want)
			}
		})
	}
}
