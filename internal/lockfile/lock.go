package lockfile

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path"
)

type LockVersion struct {
	Version string `json:"version"`
	Image   string `json:"image"`
	Message string `json:"commitMessage"`
	Date    string `json:"date"`
}

var (
	goshipDirName      = ".goship"
	goshipLockFilename = "goship-lock.json"
)

func CreateLockFile() (*os.File, error) {
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

func Read(file io.Reader) ([]LockVersion, error) {
	var data []LockVersion

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry LockVersion

		b := scanner.Bytes()

		err := json.Unmarshal(b, &entry)
		if err != nil {
			return nil, err
		}

		data = append(data, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func Write(file io.Writer, entry LockVersion) error {
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = file.Write(b)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte("\n"))
	if err != nil {
		return err
	}

	return nil
}

func VersionExists(file io.Reader, version string) (bool, error) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry LockVersion
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return false, err
		}
		if entry.Version == version {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func OpenFile() (*os.File, error) {
	lockpath, err := LockPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(lockpath)
	if err != nil {
		return nil, err
	}

	return file, nil

}

func LockPath() (string, error) {
	return lockPath(os.Getwd)
}

func lockPath(getwd func() (string, error)) (string, error) {
	cwd, err := getwd()
	if err != nil {
		return "", err
	}

	p := path.Join(cwd, goshipDirName, goshipLockFilename)
	return p, nil

}
