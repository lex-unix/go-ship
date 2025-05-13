package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"time"

	"neite.dev/go-ship/internal/command"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/exec/localexec"
	"neite.dev/go-ship/internal/exec/sshexec"
	"neite.dev/go-ship/internal/txman"
)

type Server struct {
	Addr string
}

type app struct {
	servers []Server

	txmanager txman.Service
	lexec     localexec.Service

	history         []History
	historySorted   bool
	historyFilePath string
}

type Option func(*app)

func New(txmanager txman.Service, lexec localexec.Service, options ...Option) *app {
	cfg := config.Get()
	servers := make([]Server, len(cfg.Servers))
	for i, server := range cfg.Servers {
		servers[i] = Server{
			Addr: server,
		}
	}

	a := &app{
		txmanager:       txmanager,
		lexec:           lexec,
		servers:         servers,
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

func (app *app) Deploy(ctx context.Context) error {
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

	err = app.lexec.Run(ctx, command.BuildImage(imgWithTag, cfg.Dockerfile))
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

	transactions := []txman.Transaction{
		{
			Name:         "Pull image",
			ForwardFunc:  PullImage(ctx, registryImg),
			RollbackFunc: rollbackNoop,
		},
		{
			Name:         "Stop old container",
			ForwardFunc:  StopContainer(ctx, currentContainer),
			RollbackFunc: StartContainer(ctx, currentContainer),
		},
		{
			Name:         "Start new container",
			ForwardFunc:  RunContainer(ctx, registryImg, newContainer),
			RollbackFunc: StopContainer(ctx, newContainer),
		},
		{
			Name:         "Save server state",
			ForwardFunc:  app.AppendVersion(newVersion),
			RollbackFunc: rollbackNoop,
		},
	}

	err = app.txmanager.Tx(ctx, transactions)
	if err != nil {
		return err
	}

	return nil
}

func (app *app) Rollback(ctx context.Context, version string) error {
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

	transactions := []txman.Transaction{
		{
			Name:         "Stop current container",
			ForwardFunc:  StopContainer(ctx, currentContainer),
			RollbackFunc: StartContainer(ctx, currentContainer),
		},
		{
			Name:         "Start new container",
			ForwardFunc:  StartContainer(ctx, newContainer),
			RollbackFunc: StopContainer(ctx, newContainer),
		},
		{
			Name:         "Update server state",
			ForwardFunc:  WriteToRemoteFile(ctx, app.historyFilePath, history),
			RollbackFunc: rollbackNoop,
		},
	}

	if err := app.txmanager.Tx(ctx, transactions); err != nil {
		return err
	}

	return nil
}

func (app *app) History(ctx context.Context) ([]History, error) {
	if err := app.LoadHistory(ctx); err != nil {
		return nil, err
	}
	// sort in case load implementation would change in the future, currently this is basically noop
	app.sortHistory()

	return app.history, nil
}

func rollbackNoop(_ sshexec.Service) error { return nil }
