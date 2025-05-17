package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"sort"
	"time"

	"neite.dev/go-ship/internal/command"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/exec/localexec"
	"neite.dev/go-ship/internal/exec/sshexec"
	"neite.dev/go-ship/internal/logging"
	"neite.dev/go-ship/internal/stream"
	"neite.dev/go-ship/internal/template"
	"neite.dev/go-ship/internal/txman"
)

type App struct {
	txmanager txman.Service
	lexec     localexec.Service

	history         []History
	historySorted   bool
	historyFilePath string
}

type Option func(*App)

func WithTxManager(txmanager txman.Service) Option {
	return func(a *App) {
		a.txmanager = txmanager
	}
}

func New(lexec localexec.Service, options ...Option) *App {
	a := &App{
		lexec:           lexec,
		historyFilePath: defautlHistoryFilePath,
		historySorted:   false,
	}

	for _, option := range options {
		option(a)
	}

	return a
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

type HostOutput struct {
	Host   string
	Output string
}

func (app *App) Deploy(ctx context.Context) error {
	cfg := config.Get()
	err := app.LoadHistory(ctx)
	if err != nil {
		return fmt.Errorf("failed to read history at %s: %w", app.historyFilePath, err)
	}

	// FIXME: should use commit hash for this
	currentVersion := app.LatestVersion()
	newVersion := generateRandomString(10)
	imgWithTag := fmt.Sprintf("%s:%s", cfg.Image, newVersion)
	registryImg := fmt.Sprintf("%s/%s", cfg.Registry.Server, imgWithTag)
	currentContainer := fmt.Sprintf("%s-%s", cfg.Service, currentVersion)
	newContainer := fmt.Sprintf("%s-%s", cfg.Service, newVersion)

	env := make([]string, 0)
	for k, v := range cfg.Secrets {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	err = app.lexec.Run(ctx, command.BuildImage(imgWithTag, cfg.Dockerfile, cfg.Secrets), localexec.WithEnv(env))
	if err != nil {
		return fmt.Errorf("failed to build image %s: %w", imgWithTag, err)
	}

	err = app.lexec.Run(ctx, command.TagImage(imgWithTag, registryImg))
	if err != nil {
		return fmt.Errorf("failed to tag image %s: %w", imgWithTag, err)
	}

	err = app.lexec.Run(ctx, command.PushImage(registryImg))
	if err != nil {
		return fmt.Errorf("failed to push %s to %s: %w", imgWithTag, cfg.Registry.Server, err)
	}

	rollback, err := app.txmanager.BeginTransaction(ctx, func(ctx context.Context, tx txman.Transaction) error {
		err := tx.Do(ctx, PullImage(registryImg), nil)
		if err != nil {
			return err
		}
		err = tx.Do(ctx, StopContainer(currentContainer), StartContainer(currentContainer))
		if err != nil {
			return err
		}
		err = tx.Do(ctx, RunContainer(registryImg, newContainer), StopContainer(newContainer))
		if err != nil {
			return err
		}

		err = tx.Do(ctx, app.AppendVersion(newVersion), nil)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil && cfg.Transaction.Bypass {
		if cfg.Transaction.Bypass {
			return err
		}
		if err := rollback(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (app *App) Rollback(ctx context.Context, version string) error {
	err := app.LoadHistory(ctx)
	if err != nil {
		return fmt.Errorf("failed to read history at %s: %w", app.historyFilePath, err)
	}

	found := slices.IndexFunc(app.history, func(h History) bool { return h.Version == version })
	if found < 0 {
		return fmt.Errorf("version %s does not exist", version)
	}
	app.history[found].Timestamp = time.Now()
	history, err := json.Marshal(app.history)
	if err != nil {
		return err
	}

	cfg := config.Get()
	currentVersion := app.LatestVersion()
	service := cfg.Service
	currentContainer := fmt.Sprintf("%s-%s", service, currentVersion)
	newContainer := fmt.Sprintf("%s-%s", service, version)

	rollback, err := app.txmanager.BeginTransaction(ctx, func(ctx context.Context, tx txman.Transaction) error {
		err := tx.Do(ctx, StopContainer(currentContainer), StartContainer(currentContainer))
		if err != nil {
			return err
		}
		err = tx.Do(ctx, StartContainer(newContainer), StopContainer(newContainer))
		if err != nil {
			return err
		}
		err = tx.Do(ctx, WriteToRemoteFile(app.historyFilePath, history), nil)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil && cfg.Transaction.Bypass {
		if cfg.Transaction.Bypass {
			return err
		}
		if err := rollback(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (app *App) History(ctx context.Context, sortDir string) ([]History, error) {
	if err := app.LoadHistory(ctx); err != nil {
		return nil, err
	}

	if sortDir == "asc" {
		sort.Sort(ByDateAsc(app.history))
	} else {
		sort.Sort(ByDateDesc(app.history))
	}

	return app.history, nil
}

func (app *App) ShowServiceInfo(ctx context.Context) (map[string]string, error) {
	cfg := config.Get()
	return app.showInfo(ctx, cfg.Service)
}

func (app *App) ShowProxyInfo(ctx context.Context) (map[string]string, error) {
	container := config.Get().Proxy.Name
	return app.showInfo(ctx, container)
}

func (app *App) ServiceLogs(ctx context.Context, follow bool, lines int, since string) error {
	cfg := config.Get()
	if err := app.LoadHistory(ctx); err != nil {
		return err
	}

	container := fmt.Sprintf("%s-%s", cfg.Service, app.LatestVersion())
	return app.logs(ctx, container, follow, lines, since)
}

func (app *App) ProxyLogs(ctx context.Context, follow bool, lines int, since string) error {
	container := config.Get().Proxy.Name
	return app.logs(ctx, container, follow, lines, since)
}

func (app *App) StopService(ctx context.Context) error {
	cfg := config.Get()
	if err := app.LoadHistory(ctx); err != nil {
		return err
	}
	container := fmt.Sprintf("%s-%s", cfg.Service, app.LatestVersion())
	return app.stopContainer(ctx, container)
}

func (app *App) StopProxy(ctx context.Context) error {
	container := config.Get().Proxy.Name
	return app.stopContainer(ctx, container)
}

func (app *App) StartService(ctx context.Context) error {
	cfg := config.Get()
	if err := app.LoadHistory(ctx); err != nil {
		return err
	}
	container := fmt.Sprintf("%s-%s", cfg.Service, app.LatestVersion())
	return app.startContainer(ctx, container)
}

func (app *App) StartProxy(ctx context.Context) error {
	container := config.Get().Proxy.Name
	return app.startContainer(ctx, container)
}

func (app *App) RestartService(ctx context.Context) error {
	if err := app.StopService(ctx); err != nil {
		return err
	}
	return app.StartService(ctx)
}

func (app *App) RegistryLogin(ctx context.Context) error {
	cfg := config.Get()

	registry := cfg.Registry.Server
	username := cfg.Registry.Username
	password := cfg.Registry.Password

	return app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		err := client.Run(ctx, command.RegistryLogin(registry, username, password))
		if err != nil {
			return fmt.Errorf("failed to login to registry: %s", err)
		}
		return nil
	})
}

func (app *App) RegistryLogout(ctx context.Context) error {
	err := app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		err := client.Run(ctx, command.RegistryLogout())
		if err != nil {
			return fmt.Errorf("failed to logout from registry: %s", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (app *App) CreateConfig() error {
	data, err := template.TemplateFS.ReadFile("templates/shipit.yaml")
	if err != nil {
		return err
	}
	err = os.WriteFile("shipit.yaml", data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) logs(
	ctx context.Context,
	container string,
	follow bool,
	lines int,
	since string,
) error {
	err := app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		var lineHandler stream.LineHandler = func(line []byte) {
			logging.InfoHost(client.Host(), string(line))
		}
		var streamErrHandler stream.StreamErrHandler = func(err error) {
			logging.ErrorHostf(client.Host(), "stream: %s", err)
		}

		sw := stream.New(lineHandler, streamErrHandler)
		defer sw.Close()

		err := client.Run(ctx, command.ContainerLogs(container, follow, lines, since), sshexec.WithStdout(sw))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("one or more hosts failed to stream logs: %w", err)
	}

	return nil
}

func (app *App) startContainer(ctx context.Context, container string) error {
	return app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		err := client.Run(ctx, command.StartContainer(container))
		if err != nil {
			return fmt.Errorf("failed to start container on %s: %w", client.Host(), err)
		}
		return nil
	})
}

func (app *App) stopContainer(ctx context.Context, container string) error {
	return app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		err := client.Run(ctx, command.StopContainer(container))
		if err != nil {
			return fmt.Errorf("failed to stop container on %s: %w", client.Host(), err)
		}
		return nil
	})
}

func (app *App) showInfo(ctx context.Context, container string) (map[string]string, error) {
	output := make(map[string]string)
	err := app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		var stdout bytes.Buffer
		err := client.Run(ctx, "docker ps --filter name="+container, sshexec.WithStdout(&stdout))
		if err != nil {
			return err
		}
		output[client.Host()] = stdout.String()
		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
