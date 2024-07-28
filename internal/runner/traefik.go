package runner

import "neite.dev/go-ship/internal/commands"

var DEFAULT_TRAEFIK_OPTS = map[string]string{}

func (r *runner) RunTraefik() error {
	return r.runOverSSH(commands.RunTraefik(
		r.config.Traefik.Img,
		r.config.Traefik.Labels(),
		r.config.Traefik.Args(),
	))
}
