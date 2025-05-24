package txman

import (
	"context"
	"errors"

	"github.com/lex-unix/faino/internal/exec/sshexec"
)

type Transaction interface {
	// Do executes a forward operation and registers its corresponding rollback.
	// If forwardFn returns an error, the transaction for the current host is
	// considered failed, and this error will be propagated to trigger rollback.
	// The rollbackFn will be executed if forwardFn succeeded but a later
	// operation (on this host or another) fails.
	Do(ctx context.Context, forwardFn Callback, rollbackFn Callback) error

	// Run is a convenience wrapper around Do for simple command execution.
	// It assumes a standard way to run a command via sshexec.Service.
	Run(ctx context.Context, forwardCmd string, rollbackCmd string) error
}

type transaction struct {
	client      sshexec.Service
	hostName    string
	rollbackFns []Callback
	hasFailed   bool
	err         error
}

func (tx *transaction) Do(ctx context.Context, forwardFn Callback, rollbackFn Callback) error {
	if tx.hasFailed {
		return tx.err
	}
	select {
	case <-ctx.Done():
		tx.hasFailed = true
		tx.err = errors.New("transaction cancelled")
		return tx.err
	default:
	}

	err := forwardFn(ctx, tx.client)
	if err != nil {
		tx.hasFailed = true
		tx.err = err
		return tx.err
	}

	if rollbackFn != nil {
		tx.rollbackFns = append(tx.rollbackFns, rollbackFn)
	}

	return nil
}

func (tx *transaction) Run(ctx context.Context, forwardCmd string, rollbackCmd string) error {
	var forwardFn Callback = func(ctx context.Context, client sshexec.Service) error {
		return client.Run(ctx, forwardCmd)
	}
	var rollbackFn Callback = func(ctx context.Context, client sshexec.Service) error {
		if rollbackCmd == "" {
			return nil
		}
		return client.Run(ctx, rollbackCmd)
	}
	return tx.Do(ctx, forwardFn, rollbackFn)
}
