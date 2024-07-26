package integration

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
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
	waitForHealthy(t, MAX_RETRY, WAIT_TIME)
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

func waitForHealthy(t *testing.T, maxRetry int, waitTime time.Duration) {
	for i := 0; i < maxRetry; i++ {
		out := dockerCompose(t, "ps -a | tail -n +2 | grep -v '(healthy)' | wc -l")
		out = strings.TrimSpace(out)
		if out == "0" {
			return
		}
		time.Sleep(waitTime)
	}
	t.FailNow()
}

func appResponse() *http.Response {
	res, err := http.Get("http://localhost:3000/")
	if err != nil {
		return nil
	}
	return res
}

func waitForApp(t *testing.T, maxRetry int, waitTime time.Duration) {
	total := 2
	up := 0
	for i := 0; i < maxRetry && up != total; i++ {
		res := appResponse()
		if res != nil && res.StatusCode == http.StatusOK {
			up++
		}
		time.Sleep(waitTime)
	}
	assert.Equal(t, total, up, "app failed to become ready")
}

func assertAppIsDown(t *testing.T) {
	res := appResponse()
	assert.Equal(t, http.StatusBadGateway, res.StatusCode, "expected app to be down")
}

func assertAppIsUp(t *testing.T) {
	res := appResponse()
	assert.Equal(t, http.StatusOK, res.StatusCode, "expected app to be up")
}

func assertAppVersion(t *testing.T, want, got string) {
	assert.Equal(t, want, got)
}

func getAppVersion() string {
	res := appResponse()
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	return string(body)
}
