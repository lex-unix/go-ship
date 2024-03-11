package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
)

var (
	goshipDirName      = ".goship"
	goshipLockFilename = "goship-lock.json"
)

func latestCommitHash() (string, error) {
	c := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := c.Output()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	out = bytes.TrimSpace(out)
	return string(out), nil
}
