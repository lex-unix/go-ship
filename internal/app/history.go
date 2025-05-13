package app

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"neite.dev/go-ship/internal/exec/sshexec"
	"neite.dev/go-ship/internal/logging"
	"neite.dev/go-ship/internal/txman"
)

const (
	defautlHistoryFilePath = "~/.shipit/history.json"
)

type History struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

func (app *app) LoadHistory(ctx context.Context) error {
	resultsCh := make(chan RemoteFileContent, len(app.servers))
	defer close(resultsCh)
	seq := txman.Sequence{CommandFunc: ReadRemoteFile(ctx, app.historyFilePath, resultsCh)}
	if err := app.txmanager.Run(ctx, []txman.Sequence{seq}); err != nil {
		return err
	}

	contentByHost := make(map[string][]byte)
	for range len(app.servers) {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context timeout or canceled")
		case result := <-resultsCh:
			if result.err == nil {
				contentByHost[result.host] = result.data
			} else {
				logging.ErrorHostf(result.host, "failed to read remote file %s: %s", app.historyFilePath, result.err)
			}
		}
	}

	if len(app.servers) != len(contentByHost) || len(contentByHost) == 0 {
		return fmt.Errorf("expected to read file on %d hosts, but got %d", len(app.servers), len(contentByHost))
	}

	// TODO: compare histories from hosts and choose the first one if okay
	var contents []byte
	for _, data := range contentByHost {
		contents = data
		break
	}

	return app.loadHistory(contents)
}

func (app *app) loadHistory(raw []byte) error {
	var h []History
	err := json.Unmarshal(raw, &h)
	if err != nil {
		return fmt.Errorf("corrupted history file: %w", err)
	}
	app.history = h
	app.sortHistory()
	return nil
}

func (app *app) sortHistory() {
	if app.history == nil || app.historySorted {
		return
	}

	sorted := make([]History, len(app.history))
	copy(sorted, app.history)

	slices.SortFunc(sorted, func(a, b History) int {
		if a.Timestamp.Before(b.Timestamp) {
			return -1
		}
		if a.Timestamp.After(b.Timestamp) {
			return 1
		}
		return 0
	})

	app.history = sorted
	app.historySorted = true
}

func (app *app) AppendVersion(version string) txman.Callback {
	h := History{
		Version:   version,
		Timestamp: time.Now(),
	}
	app.history = append(app.history, h)
	app.historySorted = false
	data, marshalErr := json.Marshal(app.history)

	return func(client sshexec.Service) error {
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal history: %w", marshalErr)
		}
		return client.WriteFile(app.historyFilePath, data)
	}
}

func (app *app) LatestVersion() string {
	app.sortHistory()
	historyLen := len(app.history)
	if historyLen == 0 {
		return ""
	}
	return app.history[historyLen-1].Version
}
