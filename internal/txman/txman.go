package txman

import (
	"context"
	"fmt"
	"sync"

	"neite.dev/go-ship/internal/exec/sshexec"
	"neite.dev/go-ship/internal/logging"
)

type Service interface {
	// Tx runs each transaction on a remote host.
	// If a transaction fails or a ctx is canceled, it will cancel every pending transactions and start rollback phase.
	Tx(ctx context.Context, transactions []Transaction) error

	// Run runs each sequence on a remote host.
	// In case of a command failure, it will continue execution on other hosts.
	Run(ctx context.Context, sequences []Sequence) error
}

type Callback func(client sshexec.Service) error

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

	// state holds state of each remote host, e.g. completed steps
	state map[string]*state

	mu sync.Mutex
	wg sync.WaitGroup
}

type state struct {
	host              string
	lastCompletedStep int
}

func New(conns ...sshexec.Service) *txman {
	m := &txman{
		clients: make(map[string]sshexec.Service, len(conns)),
		state:   make(map[string]*state),
		mu:      sync.Mutex{},
		wg:      sync.WaitGroup{},
	}
	for _, conn := range conns {
		m.clients[conn.Host()] = conn
	}

	return m
}

func (m *txman) RegisterHost(addr string, client sshexec.Service) {
	m.clients[addr] = client
}

func (m *txman) Tx(ctx context.Context, transactions []Transaction) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	m.initState()

	m.wg.Add(len(m.clients))
	for host, client := range m.clients {
		client := client
		state := m.state[host]

		go func() {
			defer m.wg.Done()
			for txIndex, tx := range transactions {
				select {
				case <-ctx.Done():
					return
				default:
				}

				err := tx.ForwardFunc(client)
				if err != nil {
					cancel()
					return
				}

				state.lastCompletedStep = txIndex
			}
		}()
	}

	m.wg.Wait()

	if ctx.Err() != nil {
		logging.Error("command failed on one of the servers")
		logging.Info("initiating rollback...")

		for host, client := range m.clients {
			client := client
			state := m.state[host]
			if state.lastCompletedStep == 0 {
				continue
			}
			m.wg.Add(1)
			go func() {
				defer m.wg.Done()
				for i := state.lastCompletedStep; i >= 0; i-- {
					tx := transactions[i]
					err := tx.RollbackFunc(client)
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

func (m *txman) Run(ctx context.Context, sequences []Sequence) error {
	m.wg.Add(len(m.clients))
	for host, client := range m.clients {
		client := client

		go func() {
			defer m.wg.Done()
			for _, seq := range sequences {
				select {
				case <-ctx.Done():
					return
				default:
				}

				err := seq.CommandFunc(client)
				if err != nil {
					logging.ErrorHost(host, "failed to run command")
					return
				}
			}
		}()
	}

	m.wg.Wait()

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
