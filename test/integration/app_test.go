package integration

import (
	"fmt"
	"testing"
)

func TestStartStop(t *testing.T) {
	setup(t)

	t.Log("running `goship deploy`")
	goship(t, "deploy")

	waitForApp(t, MAX_RETRY, WAIT_TIME)

	assertAppIsUp(t)

	t.Log("running `goship app stop`")
	goship(t, "app stop")

	assertAppIsDown(t)

	t.Log("running `goship app start`")
	goship(t, "app start")

	waitForApp(t, MAX_RETRY, WAIT_TIME)

	assertAppIsUp(t)
}

func TestRollback(t *testing.T) {
	setup(t)

	t.Log("running `goship deploy`")
	goship(t, "deploy")

	waitForApp(t, MAX_RETRY, WAIT_TIME)

	assertAppIsUp(t)

	want := getAppVersion()

	deployerExec(t, "git commit --amend -am \"second commit\"", "/app")
	deployerExec(t, "sh -c 'git rev-parse --short HEAD > version.txt' ", "/app")

	t.Log("running `goship redeploy`")
	goship(t, "redeploy")

	waitForApp(t, MAX_RETRY, WAIT_TIME)

	assertAppIsUp(t)

	t.Logf("running `goship rollback %s`", want[:7])
	goship(t, fmt.Sprintf("rollback %s", want[:7]))

	waitForApp(t, MAX_RETRY, WAIT_TIME)

	assertAppIsUp(t)

	assertAppVersion(t, want, getAppVersion())
}
