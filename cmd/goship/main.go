package main

import (
	"io"
	"path"

	"log"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"neite.dev/go-ship/internal/config"
)

func NewSSHClient() *ssh.Client {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting user home dir: %v\n", err)
	}
	hostkeyCallback, err := knownhosts.New(path.Join(homedir, ".ssh/known_hosts"))
	if err != nil {
		log.Fatalf("error reading known_hosts file: %v\n", err)
	}

	key, err := os.ReadFile(path.Join(homedir, ".ssh/id_rsa"))
	if err != nil {
		log.Fatalf("error reading id_rsa file: %v\n", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("error parsing private key: %v\n", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostkeyCallback,
	}

	client, err := ssh.Dial("tcp", "95.216.170.196:22", config)
	if err != nil {
		log.Fatalf("unable to connect: %v\n", err)
	}

	return client

}

var cmds = []string{
	"docker --version",
	"docker pull caps1d/go-ship",
	"docker images",
}

func main() {
	var cfg *config.UserConfig

	if _, err := os.Stat("config.yaml"); err != nil {
		config.NewConfig()
	}

	cfg = config.ReadUserConfig()

	log.Printf("Connecting to server: %v; Docker Hub credentials: %v, %v\n", cfg.SSH.Host, cfg.Registry.Username, cfg.Registry.RepoName)

	client := NewSSHClient()
	defer func() {
		err := client.Close()
		if err != nil {
			log.Fatalf("error closing client: %v\n", err)
		}
	}()

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("error creating new session: %v\n", err)
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := session.Shell(); err != nil {
		log.Fatal(err)
	}

	for i, cmd := range cmds {
		stdin.Write([]byte(cmd + "\n"))
		// gets stuck waiting here, need to check error without using session.Wait
		if err := session.Wait(); err != nil {
			if i == 0 {
				installCmds := []string{
					"sudo apt-get update",
					"sudo apt-get install ca-certificates curl",
					"sudo install -m 0755 -d /etc/apt/keyrings",
					"sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc",
					"sudo chmod a+r /etc/apt/keyrings/docker.asc",
					`echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
                    $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
                    sudo tee /etc/apt/sources.list.d/docker.list > /dev/null`,
					"sudo apt-get update",
				}
				// Execute installation commands
				for _, installCmd := range installCmds {
					stdin.Write([]byte(installCmd + "\n"))
					if err := session.Wait(); err != nil {
						log.Fatalf("Error executing installation command '%s': %v\n", installCmd, err)
					}
				}
				continue
			}
			log.Println(err)
		}
	}
	err = session.Close()
	if err != nil && err != io.EOF {
		log.Printf("error closing session: %v\n", err)
	}

}
