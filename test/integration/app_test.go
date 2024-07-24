package integration

import (
	"fmt"
	"testing"
)

func TestStartStop(t *testing.T) {
	setup(t)

	t.Log("Deploying app (running `goship deploy`)")

	goship(t, "deploy")
	waitForApp(t, MAX_RETRY, WAIT_TIME)
	assertOkResponse(t, appResponse(t))

	t.Log("Stopping container (running `goship stop`)")

	goship(t, "stop")
	if !appIsDown(t) {
		t.Fatal("`goship stop` failed to stop the container")
	}

	t.Log("Starting container (running `goship start`)")

	goship(t, "start")
	waitForApp(t, MAX_RETRY, WAIT_TIME)
	assertOkResponse(t, appResponse(t))
}

func TestRollback(t *testing.T) {
	setup(t)

	t.Log("Deploying app (running `goship deploy`)")

	goship(t, "deploy")
	waitForApp(t, MAX_RETRY, WAIT_TIME)
	assertOkResponse(t, appResponse(t))

	want := getLatestAppVersion(t)

	deployerExec(t, "git commit --amend -am \"second commit\"", "/app")
	deployerExec(t, "sh -c 'git rev-parse --short HEAD > version.txt' ", "/app")

	t.Log("Redeploying app (running `goship redeploy`)")

	goship(t, "redeploy")
	waitForApp(t, MAX_RETRY, WAIT_TIME)
	assertOkResponse(t, appResponse(t))

	t.Logf("Performing rollback (running `goship rollback %s`)", want[:7])

	goship(t, fmt.Sprintf("rollback %s", want[:7]))
	waitForApp(t, MAX_RETRY, WAIT_TIME)
	assertOkResponse(t, appResponse(t))

	got := getLatestAppVersion(t)
	if want != got {
		t.Fatalf("app version not matching: got %s, want %s", got, want)
	}
}
