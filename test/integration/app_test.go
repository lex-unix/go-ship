package integration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartStop(t *testing.T) {
	setup(t)

	t.Log("running `shipit deploy`")
	shipit(t, "deploy")

	waitForApp(t, maxRetry, waitTime)

	assertAppIsUp(t)

	t.Log("running `shipit app info`")
	info := shipit(t, "app show")

	assert.Regexp(t, regexp.MustCompile("Host: vm1"), info)
	assert.Regexp(t, regexp.MustCompile("Host: vm2"), info)

	t.Log("running `shipit app stop`")
	shipit(t, "app stop")

	assertAppIsDown(t)

	t.Log("running `shipit app start`")
	shipit(t, "app start")

	waitForApp(t, maxRetry, waitTime)

	assertAppIsUp(t)
}

func TestRollback(t *testing.T) {
	setup(t)

	t.Log("running `shipit deploy`")
	shipit(t, "deploy")

	waitForApp(t, maxRetry, waitTime)

	assertAppIsUp(t)

	want := getAppVersion()

	deployerExec(t, "git commit --amend -am \"second commit\"", "/app")
	deployerExec(t, "sh -c 'git rev-parse --short HEAD > version.txt' ", "/app")

	t.Log("running `shipit redeploy`")
	shipit(t, "redeploy")

	waitForApp(t, maxRetry, waitTime)

	assertAppIsUp(t)

	t.Logf("running `shipit rollback %s`", want[:7])
	shipit(t, fmt.Sprintf("rollback %s", want[:7]))

	waitForApp(t, maxRetry, waitTime)

	assertAppIsUp(t)

	assertAppVersion(t, want, getAppVersion())
}
