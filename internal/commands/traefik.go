package commands

import "neite.dev/go-ship/internal/config"

func RunTraefik(t *config.Traefik) string {
	return Docker(
		"run",
		"-d",
		"-p 80:80",
		"--name traefik",
		"--volume /var/run/docker.sock:/var/run/docker.sock:ro",
		t.Labels(),
		t.Img,
		"--providers.docker",
		"--entryPoints.web.address=:80",
		"--accesslog=true",
		t.Args(),
	)

}
