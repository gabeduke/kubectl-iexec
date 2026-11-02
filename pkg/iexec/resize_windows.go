//go:build windows

package iexec

import (
	"context"

	"golang.org/x/term"
	"k8s.io/client-go/tools/remotecommand"
)

// watchTerminalSize sends the initial size only: Windows has no SIGWINCH.
func watchTerminalSize(_ context.Context, fd int) remotecommand.TerminalSizeQueue {
	sizes := make(sizeQueue, 1)
	if w, h, err := term.GetSize(fd); err == nil {
		sizes <- remotecommand.TerminalSize{Width: uint16(w), Height: uint16(h)} //nolint:gosec // fits in uint16
	}
	return sizes
}
