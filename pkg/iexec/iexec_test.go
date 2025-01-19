package iexec

import (
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// Mock UI (no promptui usage)
type MockUI struct {
	SelectPodFn       func([]corev1.Pod, Config) (corev1.Pod, error)
	SelectContainerFn func([]corev1.Container, Config) (corev1.Container, error)
}

func (m *MockUI) SelectPod(pods []corev1.Pod, cfg Config) (corev1.Pod, error) {
	return m.SelectPodFn(pods, cfg)
}
func (m *MockUI) SelectContainer(containers []corev1.Container, cfg Config) (corev1.Container, error) {
	return m.SelectContainerFn(containers, cfg)
}

// Mock K8s
type MockK8sClient struct {
	FetchAllPodsFn    func(namespace, labelSelector string) (*corev1.PodList, error)
	MatchPodsFn       func(pods *corev1.PodList, config Config) (*corev1.PodList, error)
	MatchContainersFn func(pod corev1.Pod, config Config) ([]corev1.Container, error)
}

func (m *MockK8sClient) FetchAllPods(ns, lbl string) (*corev1.PodList, error) {
	return m.FetchAllPodsFn(ns, lbl)
}
func (m *MockK8sClient) MatchPods(pods *corev1.PodList, cfg Config) (*corev1.PodList, error) {
	return m.MatchPodsFn(pods, cfg)
}
func (m *MockK8sClient) MatchContainers(pod corev1.Pod, cfg Config) ([]corev1.Container, error) {
	return m.MatchContainersFn(pod, cfg)
}

// Mock Exec
type MockExecRunner struct {
	RunExecFn func(*rest.Config, corev1.Pod, corev1.Container, []string) error
}

func (m *MockExecRunner) RunExec(r *rest.Config, p corev1.Pod, c corev1.Container, cmd []string) error {
	return m.RunExecFn(r, p, c, cmd)
}

// Example Test
func TestIexec_Do_Success(t *testing.T) {
	ui := &MockUI{
		SelectPodFn: func(pods []corev1.Pod, cfg Config) (corev1.Pod, error) {
			return pods[0], nil
		},
		SelectContainerFn: func(containers []corev1.Container, cfg Config) (corev1.Container, error) {
			return containers[0], nil
		},
	}
	k8sMock := &MockK8sClient{
		FetchAllPodsFn: func(ns, lbl string) (*corev1.PodList, error) {
			return &corev1.PodList{
				Items: []corev1.Pod{{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "test-container"}},
					},
				}},
			}, nil
		},
		MatchPodsFn: func(pods *corev1.PodList, cfg Config) (*corev1.PodList, error) {
			return pods, nil
		},
		MatchContainersFn: func(pod corev1.Pod, cfg Config) ([]corev1.Container, error) {
			return pod.Spec.Containers, nil
		},
	}
	execMock := &MockExecRunner{
		RunExecFn: func(r *rest.Config, p corev1.Pod, c corev1.Container, cmd []string) error {
			return nil
		},
	}

	cfg := &Config{Namespace: "default"}
	iex := NewIexec(&rest.Config{}, cfg, ui, k8sMock, execMock)

	if err := iex.Do(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// Example Test - Error scenario
func TestIexec_Do_FetchPodsError(t *testing.T) {
	ui := &MockUI{}
	k8sMock := &MockK8sClient{
		FetchAllPodsFn: func(ns, lbl string) (*corev1.PodList, error) {
			return nil, errors.New("error fetching pods")
		},
		MatchPodsFn:       nil,
		MatchContainersFn: nil,
	}
	execMock := &MockExecRunner{}

	cfg := &Config{Namespace: "default"}
	iex := NewIexec(&rest.Config{}, cfg, ui, k8sMock, execMock)

	err := iex.Do()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "fetchAllPods failed: error fetching pods" {
		t.Errorf("unexpected error message: %v", err)
	}
}
