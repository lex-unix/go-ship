package app

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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

// ByDateAsc is a helper type for History slice that implements sort.Interface
type ByDateAsc []History

func (a ByDateAsc) Len() int           { return len(a) }
func (a ByDateAsc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDateAsc) Less(i, j int) bool { return a[i].Timestamp.Before(a[j].Timestamp) }

// ByDateDesc is a helper type for History slice that implements sort.Interface
type ByDateDesc []History

func (a ByDateDesc) Len() int           { return len(a) }
func (a ByDateDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDateDesc) Less(i, j int) bool { return a[i].Timestamp.After(a[j].Timestamp) }

func (app *App) LoadHistory(ctx context.Context) error {
	if app.history != nil {
		return nil
	}
	resultsCh := make(chan RemoteFileContent, len(app.servers))
	defer close(resultsCh)
	err := app.txmanager.Execute(ctx, ReadRemoteFile(app.historyFilePath, resultsCh))
	if err != nil {
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

func (app *App) loadHistory(raw []byte) error {
	var h []History
	err := json.Unmarshal(raw, &h)
	if err != nil {
		return fmt.Errorf("corrupted history file: %w", err)
	}
	app.history = h
	app.sortHistory()
	return nil
}

// sortHistory sorts history in descending order and modifies history slice.
// If history is empty or already sorted it does nothing.
func (app *App) sortHistory() {
	if app.history == nil || app.historySorted {
		return
	}

	sort.Sort(ByDateDesc(app.history))
	app.historySorted = true
}

func (app *App) AppendVersion(version string) txman.Callback {
	h := History{
		Version:   version,
		Timestamp: time.Now(),
	}
	app.history = append(app.history, h)
	app.historySorted = false
	data, marshalErr := json.Marshal(app.history)

	return func(ctx context.Context, client sshexec.Service) error {
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal history: %w", marshalErr)
		}
		return client.WriteFile(app.historyFilePath, data)
	}
}

func (app *App) LatestVersion() string {
	app.sortHistory()
	if app.history == nil {
		return ""
	}
	if len(app.history) == 0 {
		return ""
	}
	return app.history[0].Version
}
