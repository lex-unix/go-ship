package docker

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"neite.dev/go-ship/internal/ssh"
)

const quote = "\""

type DockerCmd struct {
	cmd string
}

func (c DockerCmd) Run() error {
	cmd := exec.Command("sh", "-c", c.cmd)
	return run(cmd)
}

func (c DockerCmd) RunSSH(client *ssh.Client) error {
	session, err := client.NewSession(ssh.WithStdout(os.Stdout), ssh.WithStderr(os.Stderr))
	if err != nil {
		return err
	}
	defer session.Close()
	return session.Run(c.cmd)
}

func IsInstalled() DockerCmd {
	return DockerCmd{cmd: "docker --version"}
}

func IsRunning() DockerCmd {
	return DockerCmd{cmd: "docker version"}
}

func BuildImage(imgName, path string) DockerCmd {
	return DockerCmd{cmd: fmt.Sprintf("docker build --platform=linux/amd64 -t %s ./test/integration/docker/app", imgName)}
}

func RunContainer(port int, name, image string) DockerCmd {
	portMap := fmt.Sprintf("%d:%d", port, port)
	return DockerCmd{cmd: fmt.Sprintf("docker run -d -p %s --name %s %s", portMap, name, image)}

}

func ListContainers() DockerCmd {
	// cmd := exec.Command("docker", "ps", "-a", "--format", "\"{{.Names}}\"")
	// out, _ := cmd.Output()
	//
	// out = bytes.TrimSpace(out)
	// list := bytes.Split(out, []byte("\n"))
	// outList := make([]string, len(list))
	// for i := range list {
	// 	container := list[i]
	// 	container = bytes.TrimPrefix(container, []byte(quote))
	// 	container = bytes.TrimSuffix(container, []byte(quote))
	// 	outList[i] = string(container)
	// }
	return DockerCmd{cmd: "docker ps -a --format \"{{.Names}}\""}
}

func ListImages() DockerCmd {
	return DockerCmd{cmd: "docker images"}
}

func LoginToHub(user, pw string) DockerCmd {
	return DockerCmd{cmd: fmt.Sprintf("docker login -u %s -p %s", user, pw)}
}

func RenameImage(image, user, repo string) DockerCmd {
	return DockerCmd{cmd: fmt.Sprintf("docker tag %s %s/%s", image, user, repo)}
}

func PushToHub(image string) DockerCmd {
	return DockerCmd{cmd: fmt.Sprintf("docker push %s", image)}
}

func PullFromHub(user, repo string) DockerCmd {
	return DockerCmd{cmd: fmt.Sprintf("docker pull %s/%s", user, repo)}
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
			// errMsg := string(exitErr.Stderr)
			return ssh.ErrExit
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
