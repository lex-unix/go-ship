package integration

import (
	"net/http"
	"testing"
)

func TestApp(t *testing.T) {
	setup(t)

	goship(t, "deploy")
	waitForApp(t, MAX_RETRY, WAIT_TIME)
	res := appResponse(t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected respone status code: got %d, want %d", res.StatusCode, http.StatusOK)
	}

	goship(t, "stop")
	if !appIsDown(t) {
		t.Fatal("`goship stop` failed to stop the container")
	}

	goship(t, "start")
	waitForApp(t, MAX_RETRY, WAIT_TIME)
	res = appResponse(t)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("`goship start` with response status code: %d", res.StatusCode)
	}
}
