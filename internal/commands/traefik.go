package commands

func RunTraefik() string {
	return Docker(
		"run",
		"-d",
		"-p 80:80",
		"--name traefik",
		"--volume /var/run/docker.sock:/var/run/docker.sock:ro",
		"--label traefik.http.routers.catchall.entryPoints=web",
		"--label traefik.http.routers.catchall.rule=\"PathPrefix(\\`/\\`)\"",
		"--label traefik.http.routers.catchall.service=unavailable",
		"--label traefik.http.routers.catchall.priority=1",
		"--label traefik.http.services.unavailable.loadbalancer.server.port=0",
		"registry:4443/traefik:v3.1",
		"--providers.docker",
		"--entryPoints.web.address=:80",
		"--accesslog=true",
	)

}
