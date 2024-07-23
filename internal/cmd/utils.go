package cmd

import (
	"bytes"
	"errors"
	"os/exec"
	"time"
)

func latestCommitHash() (string, error) {
	c := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	out = bytes.TrimSpace(out)
	return string(out), nil
}

func latestCommitMsg() (string, error) {
	c := exec.Command("sh", "-c", "git log -1 --pretty=%B")
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	parts := bytes.Split(out, []byte("\n"))
	if len(parts) == 0 {
		return "", errors.New("could not get latest commit message")
	}
	msg := bytes.TrimSpace(parts[0])
	return string(msg), nil
}

func now() string {
	return time.Now().Format("15:04:05 02-01-2006")
}
