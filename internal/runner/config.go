package runner

import (
	"fmt"

	"neite.dev/go-ship/internal/config"
)

func (r *runner) CreateConfig() error {
	if config.Exists() {
		fmt.Println("config file already exists; skipping")
		return nil
	}
	err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize configuration file")
	}

	fmt.Println("Initialized config file `goship.yaml` in the current current directory")
	return nil
}
