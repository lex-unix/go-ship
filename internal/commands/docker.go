package commands

import "strings"

func Docker(args ...string) string {
	args = append([]string{"docker"}, args...)
	return strings.Join(args, " ")
}

func IsDockerInstalled() string {
	return Docker("--version")
}
