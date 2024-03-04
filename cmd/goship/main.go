package main

import (
	"errors"
	"io"

	"log"
	"os"

	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/docker"
	"neite.dev/go-ship/internal/ssh"
)

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

	client, err := ssh.NewConnection(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := client.Close()
		if err != nil {
			log.Fatalf("error closing client: %v\n", err)
		}
	}()

	for i, cmd := range cmds {
		session, err := client.NewSession(ssh.WithStdout(os.Stdout), ssh.WithStderr(os.Stderr))
		if err != nil {
			log.Fatalf("error creating new session: %v\n", err)
		}

		if err = session.Run(cmd); err != nil {
			// handle that in session package and create custom errors
			switch {
			case errors.Is(err, ssh.ErrExit):
				if i == 0 {
					session.Close()
					// call docker install func to run shell script
					if err := docker.InstallDocker(client); err != nil {
						log.Fatal(err)
					}
					continue
				}
				log.Fatal(err)
			default:
				log.Fatal(err)
			}
		}

		if err = session.Close(); err != nil && err != io.EOF {
			log.Fatal(err)
		}
	}

	session, err := client.NewSession(ssh.WithStdout(os.Stdout))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := session.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := session.Run("exit"); err != nil {
		log.Fatal(err)
	}
}
