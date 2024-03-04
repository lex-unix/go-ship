package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

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
	RepoName string `yaml:"reponame"`
	Password string `yaml:"password"`
}

type UserConfig struct {
	Service  string    `yaml:"service"`
	Image    string    `yaml:"image"`
	Servers  []string  `yaml:"servers"`
	SSH      *SSH      `yaml:"ssh"`
	Registry *Registry `yaml:"registry"`
}

// Reads user's config file into UserConfig struct
func ReadUserConfig() *UserConfig {
	template, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	var config UserConfig

	err = yaml.Unmarshal(template, &config)
	if err != nil {
		log.Fatalln(err)
	}

	return &config
}

// Creates new config file for the user to fill out
func NewConfig() {
	template, err := os.ReadFile("./internal/config/templates/user_config.yaml")
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	dest, err := os.Create("./config.yaml")
	if err != nil {
		log.Fatalf("error creating file: %v", err)
	}

	dest.Write(template)
	dest.Seek(0, 0)
}
