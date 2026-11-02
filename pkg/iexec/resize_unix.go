//go:build !windows

package iexec

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
	"k8s.io/client-go/tools/remotecommand"
)

// watchTerminalSize sends the current size, then a new one on each SIGWINCH,
// until ctx is cancelled.
func watchTerminalSize(ctx context.Context, fd int) remotecommand.TerminalSizeQueue {
	sizes := make(sizeQueue, 1)
	send := func() {
		w, h, err := term.GetSize(fd)
		if err != nil {
			return
		}
		select {
		case sizes <- remotecommand.TerminalSize{Width: uint16(w), Height: uint16(h)}: //nolint:gosec // terminal sizes fit in uint16
		default: // a resize is already pending, drop this one
		}
	}

	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	send()
	go func() {
		defer signal.Stop(winch)
		defer close(sizes)
		for {
			select {
			case <-winch:
				send()
			case <-ctx.Done():
				return
			}
		}
	}()
	return sizes
}
