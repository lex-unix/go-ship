package integration

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

const (
	MAX_RETRY = 5
	WAIT_TIME = 5 * time.Second
)

func dockerCompose(t *testing.T, composeCmd string) {
	t.Helper()

	cmd := exec.Command("sh", "-c", fmt.Sprintf("docker compose %s", composeCmd))
	if out, err := cmd.Output(); err != nil {
		var reason []byte
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			reason = (exitErr.Stderr)
		} else {
			reason = out
		}
		t.Fatalf("command `docker compose %s` failed: %s\n%s", composeCmd, err, string(reason))
	}
}

func deployerExec(t *testing.T, cmd string, workdir string) {
	dockerCompose(t, fmt.Sprintf("exec --workdir %s deployer %s", workdir, cmd))
}

func goship(t *testing.T, cmd string) {
	deployerExec(t, fmt.Sprintf("goship %s", cmd), "/app")
}

func setup(t *testing.T) {
	t.Log("Setting up docker compose project")
	dockerCompose(t, "up -d --build")
	t.Log("Setting up deployer container")
	setupDeployer(t)
	t.Log("Setup successful")

	t.Cleanup(func() {
		dockerCompose(t, "down -t 1")
		t.Log("Cleaning up docker compose project")
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
