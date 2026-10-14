package main

import (
	"fmt"
	"os/exec"

	"github.com/AgentGGHobby/workstation-agent/source/core/backend/platform"
)

type linuxOS struct{}

var _ platform.OS = linuxOS{}

func (linuxOS) Shutdown() error {
	return run("systemctl", "poweroff")
}

func (linuxOS) Reboot() error {
	return run("systemctl", "reboot")
}

func (linuxOS) Sleep() error {
	return run("systemctl", "suspend")
}

// Lock sperrt alle Sitzungen. Auf dem Tower läuft nur ein Benutzer, und ein
// User-Service gehört nicht zur Session-Scope, deshalb wäre lock-session ohne
// Session-ID unzuverlässig.
func (linuxOS) Lock() error {
	return run("loginctl", "lock-sessions")
}

func run(name string, args ...string) error {
	if out, err := exec.Command(name, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("linux: %s: %w: %s", name, err, string(out))
	}
	return nil
}
