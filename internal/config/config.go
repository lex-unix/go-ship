package config

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"neite.dev/go-ship/internal/validator"
)

var (
	appName             = "goship"
	userConfigFilename  = "goship.yaml"
	defaultSSHPort      = 22
	defaultSSHUser      = "root"
	defaultTraefikImage = "traefik:v3.1"

	ErrNotExists = errors.New("config does not exist")
)

//go:embed templates/*
var templateFS embed.FS

type Traefik struct {
	Img       string                 `mapstructure:"image"`
	ProxyArgs map[string]interface{} `mapstructure:"args"`
	AppLabels []string               `mapstructure:"labels"`
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
	Service    string   `mapstructure:"service"`
	Image      string   `mapstructure:"image"`
	Dockerfile string   `mapstructure:"dockerfile"`
	Servers    []string `mapstructure:"servers"`
	SSH        SSH      `mapstructure:"ssh"`
	Registry   Registry `mapstructure:"registry"`
	Traefik    Traefik  `mapstructure:"traefik"`
}

func Load() (*Config, error) {
	viper.SetConfigName(appName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("ssh.port", 22)
	viper.SetDefault("ssh.user", "root")
	viper.SetDefault("traefik.image", defaultTraefikImage)
	viper.SetDefault("traefik.labels", defaultTraefikLabels)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, ErrNotExists
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// ReadConfig reads user's config file into UserConfig struct
func ReadConfig() (*Config, error) {
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

func Validate(cfg *Config) error {
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

func loadConfig(file io.Reader) (*Config, error) {
	d := yaml.NewDecoder(file)

	var config Config
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func defaultConfig(config *Config) *Config {
	if config.SSH.Port == 0 {
		config.SSH.Port = int64(defaultSSHPort)
	}

	if config.SSH.User == "" {
		config.SSH.User = defaultSSHUser
	}

	if config.Traefik.Img == "" {
		config.Traefik.Img = defaultTraefikImage
	}

	return config
}

func validateConfig(v *validator.Validator, config *Config) {
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
