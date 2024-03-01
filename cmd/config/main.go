package main

import (
	"fmt"
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
	Image    string   `yaml:"image"`
	Server   string   `yaml:"server"`
	Username []string `yaml:"username"`
	Password []string `yaml:"password"`
}

type UserConfig struct {
	Service  string    `yaml:"service"`
	Image    string    `yaml:"image"`
	Servers  []string  `yaml:"servers"`
	SSH      *SSH      `yaml:"ssh"`
	Registry *Registry `yaml:"registry"`
}

func main() {
	template, err := os.ReadFile("./cmd/config/templates/config.yaml")
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	dest, err := os.Create("./newConf.yaml")
	if err != nil {
		log.Fatalf("error creating file: %v", err)
	}

	dest.Write(template)
	dest.Seek(0, 0)

	var config UserConfig

	destContent := make([]byte, len(template))

	_, err = dest.Read(destContent)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(destContent))

	err = yaml.Unmarshal(destContent, &config)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%#v\n", config.SSH)
}
