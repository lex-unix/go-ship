package app

import (
	"context"
	"fmt"

	"neite.dev/go-ship/internal/command"
	"neite.dev/go-ship/internal/exec/sshexec"
	"neite.dev/go-ship/internal/txman"
)

func PullImage(img string) txman.Callback {
	return func(ctx context.Context, client sshexec.Service) error {
		return client.Run(ctx, command.PullImage(img))
	}
}

func RunContainer(img string, container string) txman.Callback {
	return func(ctx context.Context, client sshexec.Service) error {
		return client.Run(ctx, command.RunContainer(img, container))
	}
}

func StopContainer(containerName string) txman.Callback {
	return func(ctx context.Context, client sshexec.Service) error {
		return client.Run(ctx, command.StopContainer(containerName))
	}
}

func StartContainer(containerName string) txman.Callback {
	return func(ctx context.Context, client sshexec.Service) error {
		return client.Run(ctx, command.StartContainer(containerName))
	}
}

type RemoteFileContent struct {
	host string
	data []byte
	err  error
}

func ReadRemoteFile(path string, resultsCh chan<- RemoteFileContent) txman.Callback {
	return func(ctx context.Context, client sshexec.Service) error {
		data, err := client.ReadFile(path)
		result := RemoteFileContent{
			host: client.Host(),
			data: data,
			err:  err,
		}
		resultsCh <- result
		if err != nil {
			return fmt.Errorf("host %s: failed to read file %q: %w", client.Host(), path, err)
		}
		return nil
	}
}

func WriteToRemoteFile(path string, data []byte) txman.Callback {
	return func(ctx context.Context, client sshexec.Service) error {
		return client.WriteFile(path, data)
	}
}

func rollbackNoop(_ context.Context, _ sshexec.Service) error { return nil }
