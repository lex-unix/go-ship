package cliutil

import (
	"neite.dev/go-ship/internal/app"
	"neite.dev/go-ship/internal/exec/localexec"
	"neite.dev/go-ship/internal/txman"
)

type Factory struct {
	App       *app.App
	Txman     txman.Service
	Localexec localexec.Service
}
