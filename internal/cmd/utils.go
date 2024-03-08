package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
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

func createLockFile() (*os.File, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	lockPath := path.Join(cwd, goshipDirName)
	err = os.Mkdir(lockPath, 0755)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(path.Join(lockPath, goshipLockFilename))
	if err != nil {
		return nil, err
	}
	return f, err
}

func lockFilePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return path.Join(cwd, goshipDirName, goshipLockFilename), nil
}
