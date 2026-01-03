package main

import (
	"fmt"
	"strings"

	"github.com/HazelnutParadise/insyra/py"
)

func (a *App) PipList() (map[string]string, error) {
	return py.PipList()
}

func (a *App) PipInstall(pkg string) error {
	name := strings.TrimSpace(pkg)
	if name == "" {
		return fmt.Errorf("package name cannot be empty")
	}
	return py.PipInstall(name)
}

func (a *App) PipUninstall(pkg string) error {
	name := strings.TrimSpace(pkg)
	if name == "" {
		return fmt.Errorf("package name cannot be empty")
	}
	return py.PipUninstall(name)
}

func (a *App) ReinstallPyEnv() error {
	return py.ReinstallPyEnv()
}
