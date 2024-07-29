package runner

import "neite.dev/go-ship/internal/commands"

func (r *runner) RunTraefik() error {
	return r.runOverSSH(commands.RunTraefik(&r.config.Traefik))
}
