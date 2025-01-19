package iexec

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type SpdyExecRunner struct{}

func NewSpdyExecRunner() *SpdyExecRunner {
	return &SpdyExecRunner{}
}

// sizeQueue is a buffered channel for TerminalSize
type sizeQueue chan remotecommand.TerminalSize

// Next implements the TerminalSizeQueue interface
func (s sizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-s
	if !ok {
		return nil
	}
	return &size
}

func (r *SpdyExecRunner) RunExec(restCfg *rest.Config, pod corev1.Pod, container corev1.Container, cmd []string) error {
	client, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return errors.Wrap(err, "unable to create kubernetes client")
	}

	req := client.CoreV1().RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container.Name,
			Command:   cmd,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	log.WithField("URL", req.URL()).Debug("SPDY Exec request")

	executor, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
	if err != nil {
		return errors.Wrap(err, "unable to create SPDY executor")
	}

	// Put terminal into raw mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return errors.Wrap(err, "unable to init terminal raw mode")
	}
	defer func() {
		_ = term.Restore(fd, oldState)
	}()

	// Get terminal size
	w, h, err := term.GetSize(fd)
	if err != nil {
		log.Errorf("Error getting terminal size: %v", err)
	}
	sz := make(sizeQueue, 1)
	sz <- remotecommand.TerminalSize{Width: uint16(w), Height: uint16(h)}

	// Stream
	if err := executor.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:             os.Stdin,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
		Tty:               true,
		TerminalSizeQueue: sz,
	}); err != nil {
		return errors.Wrap(err, "SPDY Stream failed")
	}

	fmt.Println()
	return nil
}
