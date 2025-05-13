package app

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatArgs(t *testing.T) {
	args := map[string]any{
		"arg1": "val1",
		"arg2": "val2",
		"arg3": "val3",
	}

	result := formatArgs(args)

	assert.Regexp(t, regexp.MustCompile(`--arg\d=val\d --arg\d=val\d --arg\d=val\d`), result)
}

func TestFormatFlags(t *testing.T) {
	flags := map[string]any{
		"arg1": "val1",
		"arg2": "val2",
		"arg3": "val3",
	}

	result := formatFlags("flag", flags)
	assert.Regexp(t, regexp.MustCompile(`--flag arg\d=val\d --flag arg\d=val\d --flag arg\d=val\d`), result)
}
