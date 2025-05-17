package txman

import (
	"context"
	"errors"
	"sync"

	"neite.dev/go-ship/internal/exec/sshexec"
	"neite.dev/go-ship/internal/logging"
)

// Callback defines a function signature for operations executed on a remote host
// using an SSH client. It's typically used for individual, non-transactional commands
// or as the building block for forward and rollback operations within a transaction.
type Callback func(ctx context.Context, client sshexec.Service) error

// TxCallback defines the function signature for orchestrating a series of transactional
// steps on a specific host. It is provided by the user to BeginTransaction.
type TxCallback func(ctx context.Context, tx Transaction) error

// RollbackFunc defines the function signature for the globally aggregated rollback operation
// returned by BeginTransaction. Calling this function will attempt to execute all
// registered rollback steps across all relevant hosts for a transaction that failed
// or needs to be manually rolled back.
type RollbackFunc func(ctx context.Context) error

type Service interface {
	// BeginTransaction executes passed callback on each remote host in transaction.
	// If transaction succeeded, the returned error is nil and rollback function is nil or no-op
	// If a command fails or ctx is canceled, returned error is not nil and rollback function can be called
	// to perform a rollback.
	BeginTransaction(ctx context.Context, callback TxCallback) (RollbackFunc, error)

	// Execute runs a provided callback on a each remote host.
	// In case of a command failure, it will continue execution on other hosts.
	Execute(ctx context.Context, callback Callback) error
}

type txman struct {
	// clients stores connections to remote host
	clients map[string]sshexec.Service

	wg sync.WaitGroup
}

func New(conns ...sshexec.Service) *txman {
	m := &txman{
		clients: make(map[string]sshexec.Service, len(conns)),
		wg:      sync.WaitGroup{},
	}
	for _, conn := range conns {
		m.clients[conn.Host()] = conn
	}

	return m
}

func (m *txman) BeginTransaction(ctx context.Context, callback TxCallback) (RollbackFunc, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	txs := make([]*transaction, 0, len(m.clients))
	for host, client := range m.clients {
		tx := &transaction{
			client:   client,
			hostName: host,
		}
		txs = append(txs, tx)
	}

	var txErr error
	var txErrMu sync.Mutex
	for _, tx := range txs {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			err := callback(ctx, tx)
			txErrMu.Lock()
			if err != nil && txErr == nil {
				txErr = err
				cancel()
			}
			txErrMu.Unlock()
		}()
	}

	m.wg.Wait()

	rollbackFn := func(ctx context.Context) error {
		var wg sync.WaitGroup
		rollbackErrCh := make(chan error, len(txs))
		for _, tx := range txs {
			if len(tx.rollbackFns) == 0 {
				continue
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := len(tx.rollbackFns) - 1; i >= 0; i-- {
					select {
					case <-ctx.Done():
						return
					default:
					}
					rollback := tx.rollbackFns[i]
					err := rollback(ctx, tx.client)
					if err != nil {
						rollbackErrCh <- err
						return
					}
				}
			}()
		}

		go func() {
			wg.Wait()
			close(rollbackErrCh)
		}()

		var err error
		for rollbackErr := range rollbackErrCh {
			err = errors.Join(err, rollbackErr)
		}
		if err != nil {
			return err
		}
		return nil
	}

	if txErr != nil {
		return rollbackFn, txErr
	}

	return rollbackFn, nil
}

func (m *txman) Execute(ctx context.Context, callback Callback) error {
	errCh := make(chan error, len(m.clients))
	m.wg.Add(len(m.clients))
	for host, client := range m.clients {
		go func() {
			defer m.wg.Done()
			err := callback(ctx, client)
			if err != nil {
				logging.ErrorHost(host, "failed to run command")
				errCh <- err
			}
		}()
	}

	go func() {
		m.wg.Wait()
		close(errCh)
	}()

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
