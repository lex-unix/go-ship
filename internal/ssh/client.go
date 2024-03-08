package ssh

import (
	"fmt"
	"os"
	"path"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"neite.dev/go-ship/internal/config"
)

type Client struct {
	conn *ssh.Client
}

func NewConnection(cfg *config.UserConfig) (*Client, error) {
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

	addr := formatAddress(cfg.Servers[0], cfg.SSH.Port)
	conn, err := ssh.Dial("tcp", addr, config)
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

func (c *Client) NewSession(opts ...sshOption) (*Session, error) {
	var options sessionOptions
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	sshSess, err := c.conn.NewSession()
	if err != nil {
		return nil, err
	}

	if options.stdout != nil {
		sshSess.Stdout = options.stdout
	}

	if options.stderr != nil {
		sshSess.Stderr = options.stderr
	}

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
