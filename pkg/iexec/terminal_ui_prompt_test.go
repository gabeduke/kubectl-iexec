package iexec

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestPodSearcher(t *testing.T) {
	pods := []corev1.Pod{
		{Namespace: "prod", Name: "api-7d9f"},
		{Namespace: "staging", Name: "Worker-1"},
	}
	search := podSearcher(pods)

	tests := []struct {
		input string
		index int
		want  bool
	}{
		{"api", 0, true},
		{"API", 0, true},
		{"prod/api", 0, true},
		{"staging", 0, false},
		{"worker", 1, true},
		{" staging/ ", 1, true},
		{"", 1, true},
	}
	for _, tt := range tests {
		if got := search(tt.input, tt.index); got != tt.want {
			t.Errorf("podSearcher(%q, %d) = %v, want %v", tt.input, tt.index, got, tt.want)
		}
	}
}

func TestContainerSearcher(t *testing.T) {
	search := containerSearcher([]corev1.Container{{Name: "istio-proxy"}, {Name: "App"}})

	if !search("PROXY", 0) || search("app", 0) || !search("app", 1) {
		t.Error("containerSearcher should match the name case-insensitively")
	}
}
