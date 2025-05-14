package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"slices"
	"sort"
	"sync"
	"time"

	"neite.dev/go-ship/internal/command"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/exec/localexec"
	"neite.dev/go-ship/internal/exec/sshexec"
	"neite.dev/go-ship/internal/logging"
	"neite.dev/go-ship/internal/stream"
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
			ForwardFunc:  PullImage(registryImg),
			RollbackFunc: rollbackNoop,
		},
		{
			Name:         "Stop old container",
			ForwardFunc:  StopContainer(currentContainer),
			RollbackFunc: StartContainer(currentContainer),
		},
		{
			Name:         "Start new container",
			ForwardFunc:  RunContainer(registryImg, newContainer),
			RollbackFunc: StopContainer(newContainer),
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
			ForwardFunc:  StopContainer(currentContainer),
			RollbackFunc: StartContainer(currentContainer),
		},
		{
			Name:         "Start new container",
			ForwardFunc:  StartContainer(newContainer),
			RollbackFunc: StopContainer(newContainer),
		},
		{
			Name:         "Update server state",
			ForwardFunc:  WriteToRemoteFile(app.historyFilePath, history),
			RollbackFunc: rollbackNoop,
		},
	}

	if err := app.txmanager.Tx(ctx, transactions); err != nil {
		return err
	}

	return nil
}

func (app *app) History(ctx context.Context, sortDir string) ([]History, error) {
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

func (app *app) ShowAppInfo(ctx context.Context) error {
	cfg := config.Get()
	if err := app.LoadHistory(ctx); err != nil {
		return err
	}

	var output bytes.Buffer
	err := app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		var stdout bytes.Buffer
		err := client.Run(ctx, "docker ps --filter name="+cfg.Service, sshexec.WithStdout(&stdout))
		if err != nil {
			return err
		}
		_, _ = output.Write(stdout.Bytes())
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Print(output.String())

	return nil
}

type logLine struct {
	host string
	line string
}

type logWriter struct {
	out        bytes.Buffer
	mu         sync.RWMutex
	pipeWriter *io.PipeWriter
	wg         sync.WaitGroup
	lineCh     chan<- logLine
}

func newLogWriter(host string, linesCh chan<- logLine) *logWriter {
	pipeReader, pipeWriter := io.Pipe()
	w := &logWriter{
		lineCh:     linesCh,
		pipeWriter: pipeWriter,
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer pipeReader.Close()
		scanner := bufio.NewScanner(pipeReader)
		for scanner.Scan() {
			line := scanner.Text()
			w.lineCh <- logLine{host: host, line: line}
		}
		if err := scanner.Err(); err != nil {
			logging.Errorf("error from pipereader: %s", err)
		}
	}()
	return w
}

func (w *logWriter) Write(p []byte) (int, error) {
	return w.pipeWriter.Write(p)
}

func (w *logWriter) Close() error {
	return w.pipeWriter.Close()
}

type StreamProcessor struct {
	host       string
	pipeWriter *io.PipeWriter
	wg         sync.WaitGroup
	lineCh     chan logLine
}

func NewStream(host string) (*StreamProcessor, <-chan logLine) {
	pr, pw := io.Pipe()
	s := &StreamProcessor{
		host:       host,
		pipeWriter: pw,
		lineCh:     make(chan logLine, 50),
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer pr.Close()
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			s.lineCh <- logLine{host: host, line: scanner.Text()}
		}
		if err := scanner.Err(); err != nil {
			logging.ErrorHostf(s.host, "scanner error: %s", err)
		}
	}()

	return s, s.lineCh
}

func (s *StreamProcessor) Write(p []byte) (int, error) {
	return s.pipeWriter.Write(p)
}

func (s *StreamProcessor) Close() error {
	return s.pipeWriter.Close()
}

func (app *app) Logs(ctx context.Context, follow bool, lines int, since string) error {
	cfg := config.Get()
	if err := app.LoadHistory(ctx); err != nil {
		return err
	}

	container := fmt.Sprintf("%s-%s", cfg.Service, app.LatestVersion())

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
