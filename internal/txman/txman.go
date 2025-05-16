package txman

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"neite.dev/go-ship/internal/exec/sshexec"
	"neite.dev/go-ship/internal/lazy"
	"neite.dev/go-ship/internal/logging"
)

type Service interface {
	// Tx runs each transaction on a remote host.
	// If a transaction fails or a ctx is canceled, it will cancel every pending transactions and start rollback phase.
	Tx(ctx context.Context, transactions []Transaction) error

	// Execute runs a provided callback on a each remote host.
	// In case of a command failure, it will continue execution on other hosts.
	Execute(ctx context.Context, callback Callback) error

	// SetPrimaryHost restricts the execution of subsequent 'Tx' or 'Execute' calls to the specified host.
	//
	// It returns error if the provided host was not found in config file.
	SetPrimaryHost(host string) error
}

type Callback func(ctx context.Context, client sshexec.Service) error

type Transaction struct {
	Name         string
	ForwardFunc  Callback
	RollbackFunc Callback
}

type Sequence struct {
	CommandFunc Callback
}

type txman struct {
	// clients stores connections to remote host
	clients map[string]sshexec.Service

	// lazyClients stores connection functions to clients
	// that must be called during 'Tx' or 'Execute'
	lazyClients map[string]*lazy.Lazy[sshexec.Service]

	// state holds state of each remote host, e.g. completed steps
	state map[string]*state

	wg sync.WaitGroup
}

type state struct {
	host              string
	lastCompletedStep int
}

func New(conns ...sshexec.Service) *txman {
	m := &txman{
		lazyClients: make(map[string]*lazy.Lazy[sshexec.Service]),
		clients:     make(map[string]sshexec.Service, len(conns)),
		state:       make(map[string]*state),
		wg:          sync.WaitGroup{},
	}
	for _, conn := range conns {
		m.clients[conn.Host()] = conn
	}

	return m
}

func (m *txman) Tx(ctx context.Context, transactions []Transaction) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := m.connect(); err != nil {
		return err
	}

	m.initState()

	m.wg.Add(len(m.clients))
	for host, client := range m.clients {
		state := m.state[host]

		go func() {
			defer m.wg.Done()
			for txStep, tx := range transactions {
				select {
				case <-ctx.Done():
					return
				default:
				}

				err := tx.ForwardFunc(ctx, client)
				if err != nil {
					cancel()
					return
				}

				state.lastCompletedStep = txStep
			}
		}()
	}

	m.wg.Wait()

	if ctx.Err() != nil {
		logging.Error("command failed on one of the servers")
		logging.Info("initiating rollback...")

		for host, client := range m.clients {
			state := m.state[host]
			if state.lastCompletedStep == 0 {
				continue
			}
			m.wg.Add(1)
			go func() {
				defer m.wg.Done()
				for i := state.lastCompletedStep; i >= 0; i-- {
					tx := transactions[i]
					err := tx.RollbackFunc(ctx, client)
					if err != nil {
						logging.ErrorHostf(host, "failed rollback step %q: %s", tx.Name, err)
					} else {
						logging.InfoHostf(host, "completed rollback step %q", tx.Name)
					}
				}
			}()
		}

		m.wg.Wait()
		logging.Info("rollack phase completed")
		return fmt.Errorf("rolled back transaction: %w", ctx.Err())
	}

	return nil
}

func (m *txman) RegisterClient(host string, lazyClient *lazy.Lazy[sshexec.Service]) {
	m.lazyClients[host] = lazyClient
}

func (m *txman) SetPrimaryHost(host string) error {
	primaryClient, registerd := m.lazyClients[host]
	if !registerd {
		return fmt.Errorf("host %s was not found in configuration file", host)
	}
	clear(m.lazyClients)
	m.lazyClients[host] = primaryClient
	return nil
}

func (m *txman) Execute(ctx context.Context, callback Callback) error {
	if err := m.connect(); err != nil {
		return err
	}

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

func (m *txman) initState() {
	clear(m.state)
	for host := range m.clients {
		m.state[host] = &state{
			host:              host,
			lastCompletedStep: -1,
		}
	}
}

func (m *txman) connect() error {
	for host, lazyClient := range m.lazyClients {
		client, err := lazyClient.Load()
		if err != nil {
			return fmt.Errorf("faield to connect to host %s: %s", host, err)
		}
		m.clients[host] = client
	}
	return nil
}
