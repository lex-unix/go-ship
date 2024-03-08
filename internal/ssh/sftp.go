package ssh

import (
	"fmt"
	"os"

	"github.com/pkg/sftp"
)

type SFTPClient struct {
	conn *sftp.Client
}

func (c *SFTPClient) TransferFile(src, dest string) error {
	srcFile, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	destFile, err := c.conn.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := destFile.Write(srcFile); err != nil {
		return err
	}
	return nil
}

func (c *SFTPClient) MakeExecutable(path string) error {
	err := c.conn.Chmod(path, 0755)
	if err != nil {
		return fmt.Errorf("error running chmod: %w", err)
	}
	return nil
}

func (c *SFTPClient) TransferExecutable(src, dest string) error {
	err := c.TransferFile(src, dest)
	if err != nil {
		return fmt.Errorf("error transfering file: %w", err)
	}
	err = c.MakeExecutable(dest)
	if err != nil {
		return err
	}
	return nil
}

func (c *SFTPClient) Close() error {
	return c.conn.Close()
}
