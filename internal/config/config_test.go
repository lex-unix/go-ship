package config

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"neite.dev/go-ship/internal/validator"
)

func TestLoadConfig(t *testing.T) {
	data := `
service: test-app

image: test-app

servers:
  - vm1
  - vm2

registry:
  server: registry:4443
  username: root
  password: root # this is comment

traefik:
  args:
    entryPoints.web.address: ":80"
    entryPoints.websecure.address: ":443"
    entryPoints.websecure.invalid: true`

	config, err := loadConfig(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "test-app", config.Service, "expected Service to be \"test-app\"")
	assert.Equal(t, "test-app", config.Image, "expected image to be \"test-app\"")
	assert.Equal(t, "registry:4443", config.Registry.Server, "expected registry server to be \"registry:4443\"")
	assert.Equal(t, "root", config.Registry.Password, "expected registry password to be \"root\"")
	assert.Equal(t, "root", config.Registry.Username, "expected registry username to be \"root\"")
	assert.Len(t, config.Servers, 2, "expected length of servers to be 2")
	assert.IsType(t, map[string]interface{}{}, config.Traefik.ProxyArgs, "expected traefik args to be of type map[string]interface{}")
}

func TestTraefikLabels(t *testing.T) {
	data := `
traefik:
  labels:
    traefik.http.routers.testapp.entrypoints: websecure
    traefik.enable: true
    traefik.http.services.testapp.loadbalancer.server.port: 0`

	config, err := loadConfig(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	labels := config.Traefik.Labels()

	assert.Regexp(t, regexp.MustCompile(`--label traefik.http.routers.testapp.entrypoints=websecure`), labels)
	assert.Regexp(t, regexp.MustCompile(`--label traefik.enable=true`), labels)
	assert.Regexp(t, regexp.MustCompile(`--label traefik.http.services.testapp.loadbalancer.server.port=0`), labels)
}

func TestTraefikArgs(t *testing.T) {
	data := `
traefik:
  args:
    entryPoints.web.address: ":80"
    entryPoints.websecure.address: ":443"
    entryPoints.websecure.invalid: true`

	config, err := loadConfig(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	args := config.Traefik.Args()

	assert.Regexp(t, regexp.MustCompile(`--entryPoints.web.address=:80`), args)
	assert.Regexp(t, regexp.MustCompile(`--entryPoints.websecure.address=:443`), args)
	assert.Regexp(t, regexp.MustCompile(`--entryPoints.websecure.invalid=true`), args)

}

func TestConfigValidation(t *testing.T) {
	tests := map[string]struct {
		expected bool
		input    UserConfig
	}{
		`valid`: {
			expected: true,
			input: UserConfig{
				Service: "test-service",
				Image:   "test-image",
				Servers: []string{"foo", "bar"},
				Registry: Registry{
					Username: "root",
					Password: "root",
				},
			},
		},
		`invalid`: {
			expected: false,
			input: UserConfig{
				Service: "",
				Image:   "",
				Servers: []string{"", ""},
				Registry: Registry{
					Username: "",
					Password: "",
				},
			},
		},
		`semi-valid`: {
			expected: false,
			input: UserConfig{
				Service: "test-service",
				Image:   "test-image",
				Servers: []string{},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			validateConfig(v, &tt.input)
			got := v.Valid()
			assert.Equalf(t, tt.expected, got, "validateConfig(%#v) = %v, want %v", tt.input, got, tt.expected)
		})
	}

}
