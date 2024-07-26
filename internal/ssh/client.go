package ssh

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Client struct {
	conn *ssh.Client
}

func NewConnection(addr string, port int64) (*Client, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	hostkeyCallback, err := knownhosts.New(path.Join(homedir, ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}

	key, err := os.ReadFile(path.Join(homedir, ".ssh", "id_rsa"))
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostkeyCallback,
	}

	dsn := formatAddress(addr, port)
	conn, err := ssh.Dial("tcp", dsn, config)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func (c *Client) NewSFTPClient() (*SFTPClient, error) {
	client, err := sftp.NewClient(c.conn)
	if err != nil {
		return nil, err
	}
	return &SFTPClient{conn: client}, nil
}

func (c *Client) NewSession(stdout, stderr io.Writer) (*Session, error) {
	sshSess, err := c.conn.NewSession()
	if err != nil {
		return nil, err
	}

	sshSess.Stdout = os.Stdout
	sshSess.Stderr = os.Stderr

	session := Session{
		sshSess: sshSess,
	}

	return &session, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func formatAddress(ip string, port int64) string {
	return fmt.Sprintf("%s:%d", ip, port)
}
