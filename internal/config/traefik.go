package config

import (
	"fmt"
	"strconv"
	"strings"
)

var defaultTraefikLabels = []string{
	"--label traefik.http.routers.catchall.entryPoints=web",
	"--label traefik.http.routers.catchall.rule='PathPrefix(`/`)'",
	"--label traefik.http.routers.catchall.service=unavailable",
	"--label traefik.http.routers.catchall.priority=1",
	"--label traefik.http.services.unavailable.loadbalancer.server.port=0",
}

type Traefik struct {
	Img       string                 `yaml:"image"`
	ProxyArgs map[string]interface{} `yaml:"args"`
	AppLabels map[string]interface{} `yaml:"labels"`
}

func (c *Traefik) Labels() string {
	labels := collectFlags(c.AppLabels, "label")
	labels = append(labels, defaultTraefikLabels...)

	return strings.Join(labels, " ")
}

func (c *Traefik) Args() string {
	return strings.Join(collectArgs(c.ProxyArgs), " ")
}

func collectArgs(m map[string]interface{}) []string {
	args := make([]string, 0, len(m))
	for k, v := range m {
		arg := fmt.Sprintf("--%s=%s", k, toString(v))
		args = append(args, arg)
	}

	return args
}

func collectFlags(m map[string]interface{}, flag string) []string {
	args := make([]string, 0, len(m))

	for k, v := range m {
		arg := fmt.Sprintf("--%s %s=%s", flag, k, toString(v))
		args = append(args, arg)
	}

	return args
}

func toString(v interface{}) string {
	switch t := v.(type) {
	case int:
		return strconv.Itoa(t)
	case bool:
		return strconv.FormatBool(t)
	case string:
		return t
	default:
		return ""
	}
}
