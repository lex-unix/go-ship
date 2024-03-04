package docker

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/sftp"
	"neite.dev/go-ship/internal/ssh"
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
	cmd.Env = append(cmd.Env, "DOCKER_DEFAULT_PLATFORM=linux/amd64")
	return run(cmd)
}

func RunContainer(port int, name, image string) error {
	portMap := fmt.Sprintf("%d:%d", port, port)
	cmd := exec.Command("docker", "run", "-d", "-p", portMap, "--name", name, image)
	return run(cmd)
}

func StartContainer(name string) error {
	cmd := exec.Command("docker", "start", name)
	return run(cmd)
}

func ListAllContainers() []string {
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

func LoginToHub(user, pw string) error {
	cmd := exec.Command("docker", "login", "-u", user, "-p", pw)
	return run(cmd)
}

func RenameImage(image, user, repo string) error {
	cmd := exec.Command("docker", "tag", image, user+"/"+repo)
	return run(cmd)
}

func PushToHub(user, repo string) error {
	cmd := exec.Command("docker", "push", user+"/"+repo)
	return run(cmd)
}

func InstallDocker(c *ssh.Client) error {
	// create new client for sftp
	client, err := sftp.NewClient(c.Conn)
	if err != nil {
		return err
	}
	defer client.Close()

	srcFile, err := os.Open("./scripts/setup.sh")
	if err != nil {
		return err
	}

	// create script file for sftp client
	f, err := client.Create("setup.sh")
	if err != nil {
		return err
	}
	// read the script file contents
	content, err := io.ReadAll(srcFile)
	if err != nil {
		return err
	}

	// write the content to sftp file
	if _, err := f.Write(content); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	if err := client.Chmod("setup.sh", 0755); err != nil {
		return err
	}

	// start new ssh session on our ssh.Client
	session, err := c.NewSession(ssh.WithStdout(os.Stdout))
	if err != nil {
		return err
	}

	defer session.Close()

	// execute install docker script
	if err := session.Run("./setup.sh"); err != nil {
		return err
	}

	return nil
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
