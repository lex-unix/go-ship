package config

import (
	"embed"
	"io"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

var (
	userConfigFilename = "goship.yaml"
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
		return nil, err
	}

	userConfig, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer userConfig.Close()

	config, err := loadConfig(userConfig)

	if config.Dockerfile == "" {
		config.Dockerfile = "."
	}
	if config.SSH.User == "" {
		config.SSH.User = "root"
	}
	if config.SSH.Port == 0 {
		config.SSH.Port = 22
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
		return err
	}

	dest, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = dest.Write(template)
	if err != nil {
		return err
	}

	return nil
}

// IsExists checks if the config file exists
func IsExists() bool {
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
