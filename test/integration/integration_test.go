package integration

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	MAX_RETRY = 5
	WAIT_TIME = 5 * time.Second
)

func dockerCompose(t *testing.T, composeCmd string) string {
	t.Helper()

	var buff bytes.Buffer

	cmd := exec.Command("sh", "-c", fmt.Sprintf("docker compose %s", composeCmd))
	cmd.Stderr = &buff
	cmd.Stdout = &buff
	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			t.Fatalf("command `docker compose %s` failed: %s\n%s", composeCmd, err, buff.String())
		} else {
			t.Fatalf("command `docker compose %s` failed: %s\n%s", composeCmd, err, buff.String())
		}
	}

	return buff.String()
}

func deployerExec(t *testing.T, cmd string, workdir string) string {
	return dockerCompose(t, fmt.Sprintf("exec --workdir %s deployer %s", workdir, cmd))
}

func goship(t *testing.T, cmd string) string {
	return deployerExec(t, fmt.Sprintf("/usr/local/bin/goship %s", cmd), "/app")
}

func setup(t *testing.T) {
	t.Log("Setting up docker compose project")
	dockerCompose(t, "up -d --build")
	t.Log("Setting up deployer container")
	setupDeployer(t)
	t.Log("Setup successful")

	t.Cleanup(func() {
		t.Log("Cleaning up docker compose project")
		dockerCompose(t, "down -t 1")
	})
}

func setupDeployer(t *testing.T) {
	deployerExec(t, "./setup.sh", "/")
}

func appResponse(t *testing.T) *http.Response {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Get("http://localhost:3001/")
	if err != nil {
		t.Logf("appResponse() failed: %s", err)
		return nil
	}
	return res
}

func waitForApp(t *testing.T, maxRetry int, waitTime time.Duration) {
	for i := 0; i < maxRetry; i++ {
		t.Logf("Attempt %d of %d", i+1, maxRetry)
		res := appResponse(t)
		if res != nil && res.StatusCode == http.StatusOK {
			t.Logf("App is ready")
			return
		}
		t.Logf("App not ready, waiting %s", waitTime)
		time.Sleep(waitTime)
	}

	t.Fatal("App failed to become ready")
}

func appIsDown(_ *testing.T) bool {
	return true
}

func getLatestAppVersion(t *testing.T) string {
	res := appResponse(t)
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	return string(body)
}

func assertOkResponse(t *testing.T, res *http.Response) {
	assert.True(t, res.StatusCode == http.StatusOK, "response status code != 200")
}
