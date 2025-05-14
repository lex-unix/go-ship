package localexec

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"neite.dev/go-ship/internal/logging"
)

type Service interface {
	Run(ctx context.Context, cmd string) error
}

type Command struct{}

func New() Command {
	return Command{}
}

func (c Command) Run(ctx context.Context, cmd string) error {
	command := exec.CommandContext(ctx, "sh", "-c", cmd)

	var wg sync.WaitGroup
	stdout, _ := command.StdoutPipe()
	stderr, _ := command.StderrPipe()

	var read = func(r io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			logging.Debug(line)
		}
		if err := scanner.Err(); err != nil {
			logging.Errorf("faild to read ouput: %s", err)
		}
	}

	wg.Add(2)
	go read(stdout)
	go read(stderr)

	logging.Infof("running command %q", cmd)
	if err := command.Start(); err != nil {
		return fmt.Errorf("failed to start command: %q: %w", cmd, err)
	}

	waitErr := command.Wait()
	wg.Wait()

	if waitErr != nil {
		return fmt.Errorf("failed to execute local command %s: %w", cmd, waitErr)
	}

	return nil
}
