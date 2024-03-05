package docker

import (
	"bufio"
	// "bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"neite.dev/go-ship/internal/ssh"
)

const quote = "\""

type dockerCmd struct {
	cmd string
}

func (c dockerCmd) Run() error {
	cmd := exec.Command("sh", "-c", c.cmd)
	cmd.Env = append(cmd.Env, "DOCKER_DEFAULT_PLATFORM=linux/amd64")
	return run(cmd)
}

func (c dockerCmd) RunSSH(client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	err = session.Run(c.cmd)
	if err != nil {
		return err
	}
	return nil
}

func IsInstalled() dockerCmd {
	return dockerCmd{cmd: "docker --version"}
}

func IsRunning() dockerCmd {
	return dockerCmd{cmd: "docker version"}
}

func BuildImage(name, path string) dockerCmd {
	return dockerCmd{cmd: "docker build -t goship-app-test ./test/integration/docker/app"}
}

func RunContainer(port int, name, image string) dockerCmd {
	portMap := fmt.Sprintf("%d:%d", port, port)
	return dockerCmd{cmd: fmt.Sprintf("docker run -d -p %s --name %s %s", portMap, name, image)}

}

func ListContainers() dockerCmd {
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
	return dockerCmd{cmd: "docker ps -a --format \"{{.Names}}\""}
}

func LoginToHub(user, pw string) dockerCmd {
	return dockerCmd{cmd: fmt.Sprintf("docker login -u %s -p %s", user, pw)}
}

func RenameImage(image, user, repo string) dockerCmd {
	return dockerCmd{cmd: fmt.Sprintf("docker tag %s %s/%s", image, user, repo)}
}

func PushToHub(user, repo string) dockerCmd {
	return dockerCmd{cmd: fmt.Sprintf("docker push %s/%s", user, repo)}
}

func InstallDocker(c *ssh.Client) error {
	// create new client for sftp
	client, err := c.NewSFTPClient()
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
	session, err := c.NewSession(ssh.WithStdout(os.Stdout), ssh.WithStderr(os.Stderr))
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
