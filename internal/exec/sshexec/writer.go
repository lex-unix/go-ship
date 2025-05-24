package sshexec

import (
	"bytes"

	"github.com/lex-unix/faino/internal/logging"
)

var newline = []byte("\n")

// logWriter is used to get stdout and stderr output from ssh session if no other writers were provided.
// / It usess logging.DebugHost internally
type logWriter struct {
	host string
}

// Write calls DebugHost removing leading new line from p.
// Returned error is always nill and n is always zero.
func (w *logWriter) Write(p []byte) (n int, err error) {
	logline := bytes.TrimSuffix(p, newline)
	logging.DebugHost(w.host, string(logline))
	return
}
