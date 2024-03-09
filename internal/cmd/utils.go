package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func versionExists(filePath, version string) (bool, error) {
	// Read the existing file from the file
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	data := make([]map[string]string, 0)

	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return false, err
		}
		var row map[string]string
		if err := json.Unmarshal(line, &row); err != nil {
			return false, err
		}

		data = append(data, row)
	}

	if lastObj := data[len(data)-1]; lastObj["version"] == version {
		return true, nil
	}

	return false, nil
}

func writeToLockFile(f *os.File, data map[string]string) error {
	datajson, err := json.Marshal(data)
	if err != nil {
		log.Println("error json.Marshal()")
		return err
	}
	_, err = f.Write(datajson)
	if err != nil {
		log.Println("could write to file")
		return err
	}
	_, err = f.Write([]byte(string("\n")))
	if err != nil {
		log.Println("could write to file")
		return err
	}
	return nil
}
