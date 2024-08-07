package ssh

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var (
	ErrExit = errors.New("command exit error")
)

type Client struct {
	conn *ssh.Client
	host string
}

func NewConnection(host string, port int64) (*Client, error) {
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

	dsn := formatAddress(host, port)
	conn, err := ssh.Dial("tcp", dsn, config)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn, host: host}, nil
}

func (c *Client) NewSFTPClient() (*SFTPClient, error) {
	client, err := sftp.NewClient(c.conn)
	if err != nil {
		return nil, err
	}
	return &SFTPClient{conn: client}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Address() string {
	return c.conn.RemoteAddr().String()
}

func (c *Client) Exec(cmd string) error {
	sess, err := c.conn.NewSession()
	if err != nil {
		return err
	}

	defer sess.Close()

	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	if err := sess.Run(cmd); err != nil {
		var exitErr *ssh.ExitError
		switch {
		case errors.As(err, &exitErr):
			return ErrExit
		default:
			return err
		}
	}

	return nil
}

func (c *Client) ExecWithHost(cmd string) error {
	sess, err := c.conn.NewSession()
	if err != nil {
		return err
	}

	defer sess.Close()

	fmt.Fprintf(os.Stdout, "Host: %s\n", c.host)

	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	if err := sess.Run(cmd); err != nil {
		var exitErr *ssh.ExitError
		switch {
		case errors.As(err, &exitErr):
			return ErrExit
		default:
			return err
		}
	}

	return nil
}

func formatAddress(ip string, port int64) string {
	return fmt.Sprintf("%s:%d", ip, port)
}
