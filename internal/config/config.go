package config

import (
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"neite.dev/go-ship/internal/validator"
)

// Global config instance
var cfg *Config

// default values
const (
	appName               = "shipit"
	defaultSSHPort        = 22
	defaultSSHUser        = "root"
	defaultProxyName      = "traefik"
	defaultProxyImage     = "traefik:v3.1"
	defaultRegistryServer = "docker.io"
)

var defaultProxyLabels = map[string]any{
	"traefik.http.routers.catchall.entryPoints":                  "web",
	"traefik.http.routers.catchall.rule":                         "'PathPrefix(`/`)'",
	"traefik.http.routers.catchall.service":                      "unavailable",
	"traefik.http.routers.catchall.priority":                     1,
	"traefik.http.services.unavailable.loadbalancer.server.port": 0,
}

// Config errors
var (
	ErrNotExists = errors.New("config does not exist")
)

//go:embed templates/*
var templateFS embed.FS

type Proxy struct {
	Name   string         `mapstructure:"name"`
	Img    string         `mapstructure:"image"`
	Args   map[string]any `mapstructure:"args"`
	Labels map[string]any `mapstructure:"labels"`
}

type SSH struct {
	User string `mapstructure:"user"`
	Port int64  `mapstructure:"port"`
}

type Registry struct {
	Server   string `mapstructure:"server"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Config struct {
	Service    string            `mapstructure:"service"`
	Image      string            `mapstructure:"image"`
	Dockerfile string            `mapstructure:"dockerfile"`
	Servers    []string          `mapstructure:"servers"`
	SSH        SSH               `mapstructure:"ssh"`
	Registry   Registry          `mapstructure:"registry"`
	Proxy      Proxy             `mapstructure:"proxy"`
	Debug      bool              `mapstructure:"debug"`
	Secrets    map[string]string `mapstructure:"secrets"`
}

func Load() error {
	viper.SetConfigName(appName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("ssh.port", 22)
	viper.SetDefault("ssh.user", "root")
	viper.SetDefault("proxy.name", defaultProxyName)
	viper.SetDefault("proxy.image", defaultProxyImage)
	viper.SetDefault("proxy.labels", defaultProxyLabels)
	viper.SetDefault("dockerfile", ".")
	viper.SetDefault("registry.server", defaultRegistryServer)
	viper.SetDefault("debug", false)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return ErrNotExists
		} else {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	for k, v := range cfg.Secrets {
		if expanded := os.ExpandEnv(v); expanded != "" {
			cfg.Secrets[k] = expanded
		}
	}

	if traefikLabels := viper.GetStringMap("traefik.labels"); viper.IsSet("traefik.labels") {
		mergeMaps(traefikLabels, defaultProxyLabels)
		cfg.Proxy.Labels = traefikLabels
	}

	return nil
}

func Validate() error {
	if cfg == nil {
		return errors.New("config not loaded")
	}

	v := validator.New()

	v.Check(cfg.Service != "", "service", "must include service name")
	v.Check(cfg.Image != "", "image", "must include name of the image")
	v.Check(len(cfg.Servers) > 0, "servers", "must provide at leat 1 destination server")
	v.Check(cfg.Registry.Username != "", "registry.username", "must provide registry username")
	v.Check(cfg.Registry.Password != "", "registry.password", "must provide registry password")

	if !v.Valid() {
		return v
	}

	return nil
}

func Get() *Config {
	return cfg
}

// mergeMaps merges two maps together. If key from src is present in dst, it will preserve value.
func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if _, ok := dst[k]; !ok {
			dst[k] = v
		}
	}
}
