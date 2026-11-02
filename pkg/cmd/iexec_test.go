package cmd

import (
	"reflect"
	"testing"
)

func TestRemoteCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"no args", nil, defaultShell},
		{"pod filter only", []string{"api"}, defaultShell},
		{"explicit command", []string{"api", "cat", "/etc/hosts"}, []string{"cat", "/etc/hosts"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := remoteCommand(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("remoteCommand(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
