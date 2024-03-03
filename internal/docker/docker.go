package docker

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

const quote = "\""

func IsInstalled() error {
	cmd := exec.Command("docker", "--version")
	return run(cmd)
}

func IsRunning() error {
	cmd := exec.Command("docker", "version")
	return run(cmd)
}

func BuildImage() error {
	cmd := exec.Command("docker", "build", "-t", "goship-app-test", "./test/integration/docker/app")
	return run(cmd)
}

func RunContainer(port int, name, image string) error {
	portMap := fmt.Sprintf("%d:%d", port, port)
	cmd := exec.Command("docker", "run", "-d", "-p", portMap, "--name", name, image)
	return run(cmd)
}

func ListContainers() []string {
	cmd := exec.Command("docker", "ps", "-a", "--format", "\"{{.Names}}\"")
	out, _ := cmd.Output()

	out = bytes.TrimSpace(out)
	list := bytes.Split(out, []byte("\n"))
	outList := make([]string, len(list))
	for i := range list {
		container := list[i]
		container = bytes.TrimPrefix(container, []byte(quote))
		container = bytes.TrimSuffix(container, []byte(quote))
		outList[i] = string(container)
	}
	return outList
}

func run(cmd *exec.Cmd) error {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	errChan := make(chan error)
	go readOut(stderr, errChan)

	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		switch {
		case errors.As(err, &exitErr):
			errMsg := string(exitErr.Stderr)
			return fmt.Errorf("%s", errMsg)
		default:
			return err
		}
	}

	err = <-errChan
	if err != nil {
		return err
	}

	return nil

}

func readOut(out io.Reader, errChan chan<- error) {
	r := bufio.NewReader(out)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				errChan <- err
			}
			break
		}
		fmt.Print(line)
	}

	close(errChan)
}
