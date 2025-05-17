package cliutil

import "github.com/spf13/cobra"

const skipConfigLoadAnnotation = "skipConfigLoad"

func DisableConfigLoading(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}

	cmd.Annotations[skipConfigLoadAnnotation] = "true"
}

func IsConfigLoadingEnabled(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return false
	}

	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations[skipConfigLoadAnnotation] == "true" {
			return false
		}
	}

	return true
}
