package config

import (
	"embed"
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
	User    string `yaml:"user"`
	SSHPKey string `yaml:"sshPKey"`
	Host    string `yaml:"host"`
	Port    int64  `yaml:"port"`
}

type Registry struct {
	Image    string `yaml:"image"`
	Server   string `yaml:"server"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Reponame string `yaml:"reponame"`
}

type UserConfig struct {
	Service  string    `yaml:"service"`
	Servers  []string  `yaml:"servers"`
	SSH      *SSH      `yaml:"ssh"`
	Registry *Registry `yaml:"registry"`
}

// ReadConfig reads user's config file into UserConfig struct
func ReadConfig() (*UserConfig, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	template, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config UserConfig

	err = yaml.Unmarshal(template, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
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

	dest.Write(template)

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
