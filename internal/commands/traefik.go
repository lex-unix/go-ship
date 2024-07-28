package commands

func RunTraefik(img, labels, args string) string {
	return Docker(
		"run",
		"-d",
		"-p 80:80",
		"--name traefik",
		"--volume /var/run/docker.sock:/var/run/docker.sock:ro",
		labels,
		img,
		"--providers.docker",
		"--entryPoints.web.address=:80",
		"--accesslog=true",
		args,
	)

}
