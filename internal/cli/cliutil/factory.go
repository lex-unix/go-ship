package cliutil

import (
	"fmt"
	"slices"

	"neite.dev/go-ship/internal/app"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/exec/localexec"
	"neite.dev/go-ship/internal/exec/sshexec"
	"neite.dev/go-ship/internal/txman"
)

func New() *Factory {
	f := &Factory{
		Config: configFunc(),
	}

	f.Txman = txManFunc(f)
	f.App = appFunc(f)

	return f
}

type Factory struct {
	Config func() (*config.Config, error)
	Txman  func() (txman.Service, error)
	App    func() (*app.App, error)
}

func configFunc() func() (*config.Config, error) {
	var cachedConfig *config.Config
	var configErr error
	return func() (*config.Config, error) {
		if cachedConfig != nil || configErr != nil {
			return cachedConfig, configErr
		}
		cachedConfig, configErr = config.Load()
		return cachedConfig, configErr
	}
}

func txManFunc(f *Factory) func() (txman.Service, error) {
	return func() (txman.Service, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		var hosts []string
		if cfg.Host != "" {
			found := slices.Index(cfg.Servers, cfg.Host)
			if found < 0 {
				return nil, fmt.Errorf("host %s was not found in 'servers' array", cfg.Host)
			}
			hosts = append(hosts, cfg.Host)
		} else {
			hosts = append(hosts, cfg.Servers...)
		}

		var clients []sshexec.Service
		for _, host := range hosts {
			sshClient, err := sshexec.New(host, cfg.SSH.User, cfg.SSH.Port)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to host %s: %s", host, err)
			}
			clients = append(clients, sshClient)
		}

		return txman.New(clients...), nil
	}
}

func appFunc(f *Factory) func() (*app.App, error) {
	return func() (*app.App, error) {
		txman, err := f.Txman()
		if err != nil {
			return nil, err
		}
		le := localexec.New()
		return app.New(le, app.WithTxManager(txman)), nil
	}
}
