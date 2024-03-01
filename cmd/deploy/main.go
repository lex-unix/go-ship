package main

import (
	"io"
	"path"

	"log"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func readConfig() {

}

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

func main() {
	client := NewSSHClient()
	defer func() {
		err := client.Close()
		if err != nil {
			log.Fatalf("error closing client: %v\n", err)
		}
	}()

	for range 3 {
		session, err := client.NewSession()
		if err != nil {
			log.Fatalf("error creating new session: %v\n", err)
		}

		session.Stdout = os.Stdout

		if err := session.Run("/root/prog.sh"); err != nil {
			log.Printf("error running command: %v\n", err)
			continue
		}

		err = session.Close()
		if err != nil && err != io.EOF {
			log.Printf("error closing session: %v\n", err)
		}

	}
}
