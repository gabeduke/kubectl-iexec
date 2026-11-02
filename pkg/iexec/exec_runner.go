package iexec

import (
	"context"
	"net/url"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/streaming/pkg/httpstream"
)

// StreamExecRunner runs the exec over WebSocket and falls back to SPDY when
// the API server or an intermediate proxy refuses the upgrade, like kubectl.
type StreamExecRunner struct{}

func NewStreamExecRunner() *StreamExecRunner {
	return &StreamExecRunner{}
}

// sizeQueue is a channel of TerminalSize, fed by the resize watcher.
type sizeQueue chan remotecommand.TerminalSize

// Next implements remotecommand.TerminalSizeQueue.
func (s sizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-s
	if !ok {
		return nil
	}
	return &size
}

func newExecutor(restCfg *rest.Config, u *url.URL) (remotecommand.Executor, error) {
	spdyExec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", u)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create SPDY executor")
	}
	wsExec, err := remotecommand.NewWebSocketExecutor(restCfg, "POST", u.String())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create WebSocket executor")
	}
	return remotecommand.NewFallbackExecutor(wsExec, spdyExec, httpstream.IsUpgradeFailure)
}

func (r *StreamExecRunner) RunExec(restCfg *rest.Config, pod corev1.Pod, container corev1.Container, cmd []string) error {
	client, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return errors.Wrap(err, "unable to create kubernetes client")
	}

	// Upstream #78: a remote TTY only when stdin AND stdout are terminals.
	// A TTY on `iexec pod cmd | sort` writes CRLF and escapes into the pipe.
	fd := int(os.Stdin.Fd())
	tty := term.IsTerminal(fd) && term.IsTerminal(int(os.Stdout.Fd()))

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
			Stderr:    !tty,
			TTY:       tty,
		}, scheme.ParameterCodec)

	log.WithField("URL", req.URL()).Debug("Exec request")

	executor, err := newExecutor(restCfg, req.URL())
	if err != nil {
		return err
	}

	opts := remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Tty:    tty,
	}
	if !tty {
		opts.Stderr = os.Stderr
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if tty {
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return errors.Wrap(err, "unable to init terminal raw mode")
		}
		defer func() {
			_ = term.Restore(fd, oldState)
		}()
		opts.TerminalSizeQueue = watchTerminalSize(ctx, fd)
	}

	if err := executor.StreamWithContext(ctx, opts); err != nil {
		return errors.Wrap(err, "exec stream failed")
	}
	return nil
}
