package command

import (
	"fmt"
	"strings"
)

func BuildImage(img, dockerfile string) string {
	return fmt.Sprintf("docker build -t %s %s", img, dockerfile)
}

func TagImage(img, registryImg string) string {
	return fmt.Sprintf("docker tag %s %s", img, registryImg)
}

func PushImage(img string) string {
	return fmt.Sprintf("docker push %s", img)
}

func PullImage(img string) string {
	return fmt.Sprintf("docker pull %s", img)
}

func StartContainer(img string) string {
	return fmt.Sprintf("docker start %s", img)
}

func RunContainer(img, container string) string {
	labels := []string{"--label traefik.enable=true", "--label traefik.http.routers.myapp.entrypoints=web", "--label traefik.http.routers.myapp.rule='PathPrefix(`/`)'"}
	return fmt.Sprintf("docker run -d %s --name %s %s", strings.Join(labels, " "), container, img)
}

func StopContainer(container string) string {
	return fmt.Sprintf("docker stop %s || true", container)
}

func LoginToRegistry(user, password, registry string) string {
	return fmt.Sprintf("docker login -u %s -p %s %s", user, password, registry)
}

func StartProxy(img, labels, args string) string {
	return fmt.Sprintf(
		"docker run -d -p 80:80 --name traefik --volume /var/run/docker.sock:/var/run/docker.sock:ro %s %s --providers.docker --entryPoints.web.address=:80 --accesslog=true %s",
		labels,
		img,
		args,
	)
}

func ListRunningContainers() string {
	return "docker ps"
}

func ListAllContainers() string {
	return "docker ps -a"
}

func ContainerLogs(container string, follow bool, lines int, since string) string {
	var sb strings.Builder
	sb.WriteString("docker logs")

	if len(since) != 0 {
		sb.WriteString(" --since ")
		sb.WriteString(since)
	}
	if lines != 0 {
		sb.WriteString(fmt.Sprintf(" --tail %d", lines))
	}
	if follow {
		sb.WriteString(" --follow")
	}

	sb.WriteString(" ")
	sb.WriteString(container)

	return sb.String()
}
