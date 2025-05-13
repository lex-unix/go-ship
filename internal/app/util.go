package app

import (
	"fmt"
	"strings"
)

func formatArg(k string, v any) string {
	return fmt.Sprintf("--%s=%v", k, v)
}

func formatArgs(argmap map[string]any) string {
	args := make([]string, 0, len(argmap))
	for k, v := range argmap {
		args = append(args, formatArg(k, v))
	}
	return strings.Join(args, " ")
}

func formatFlag(f, k string, v any) string {
	return fmt.Sprintf("--%s %s=%v", f, k, v)
}

func formatFlags(f string, flagmap map[string]any) string {
	flags := make([]string, 0, len(flagmap))
	for k, v := range flagmap {
		flags = append(flags, formatFlag(f, k, v))
	}
	return strings.Join(flags, " ")
}
