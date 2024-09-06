package config

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
	"neite.dev/go-ship/internal/validator"
)

var (
	userConfigFilename = "goship.yaml"
	defaultSSHPort     = 22
	defaultSSHUser     = "root"
	defaultTraefikImg  = "traefik:v3.1"
)

//go:embed templates/*
var templateFS embed.FS

type SSH struct {
	User string `yaml:"user"`
	Port int64  `yaml:"port"`
}

type Registry struct {
	Server   string `yaml:"server"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type UserConfig struct {
	Service    string   `yaml:"service"`
	Image      string   `yaml:"image"`
	Dockerfile string   `yaml:"dockerfile"`
	Servers    []string `yaml:"servers"`
	SSH        SSH      `yaml:"ssh"`
	Registry   Registry `yaml:"registry"`
	Traefik    Traefik  `yaml:"traefik"`
}

// ReadConfig reads user's config file into UserConfig struct
func ReadConfig() (*UserConfig, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer configFile.Close()

	config, err := loadConfig(configFile)
	config = defaultConfig(config)

	v := validator.New()
	if validateConfig(v, config); !v.Valid() {
		return nil, formatErrors(v.Errors)
	}
	return config, nil
}

// NewConfig creates new config file from a template
func NewConfig() error {
	template, err := templateFS.ReadFile("templates/user_config.yaml")
	if err != nil {
		return err
	}

	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to create new config file: %w", err)
	}

	dest, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create new config file: %w", err)
	}
	defer dest.Close()

	_, err = dest.Write(template)
	if err != nil {
		return fmt.Errorf("failed to write into new config file: %w", err)
	}

	return nil
}

// Exists checks if the config file exists
func Exists() bool {
	configPath, err := getConfigPath()
	if err != nil {
		return false
	}
	if _, err := os.Stat(configPath); err != nil {
		return false
	}

	return true
}

func getConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	configPath := path.Join(cwd, userConfigFilename)
	return configPath, nil
}

func loadConfig(file io.Reader) (*UserConfig, error) {
	d := yaml.NewDecoder(file)

	var config UserConfig
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func defaultConfig(config *UserConfig) *UserConfig {
	if config.SSH.Port == 0 {
		config.SSH.Port = int64(defaultSSHPort)
	}

	if config.SSH.User == "" {
		config.SSH.User = defaultSSHUser
	}

	if config.Traefik.Img == "" {
		config.Traefik.Img = defaultTraefikImg
	}

	return config
}

func validateConfig(v *validator.Validator, config *UserConfig) {
	v.Check(config.Service != "", "service", "Service was not provided")
	v.Check(config.Image != "", "image", "Image was not provided")
	v.Check(len(config.Servers) > 0, "servers", "No servers was provided")
	v.Check(config.Registry.Username != "", "registry.username", "No registry username was provided")
	v.Check(config.Registry.Password != "", "registry.password", "No registry password was provided")
}

func formatErrors(errs map[string]string) error {
	var msg strings.Builder
	for _, v := range errs {
		fmt.Fprintf(&msg, "%s\n", v)
	}
	return errors.New(msg.String())

}
