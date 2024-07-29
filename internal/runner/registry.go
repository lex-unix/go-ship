package runner

import "neite.dev/go-ship/internal/commands"

func (r *runner) RegistryLogin() error {
	cmd := commands.Docker("login", "-u", r.config.Registry.Username, "-p", r.config.Registry.Password, r.config.Registry.Server)
	err := r.runOverSSH(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (r *runner) RegistryLogout() error {
	if err := r.runOverSSH(commands.Docker("logout", r.config.Registry.Server)); err != nil {
		return err
	}
	return nil
}
