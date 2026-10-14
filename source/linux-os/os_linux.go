package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

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

// Lock sperrt die grafische Sitzung des Benutzers, unter dem der Agent läuft.
// lock-sessions (alle Sitzungen) verlangt eine polkit-Authentifizierung, das Sperren
// der eigenen Sitzung nicht. Ein User-Service hat keine eigene Session-ID, deshalb
// wird sie über die Display-Eigenschaft des Benutzers ermittelt.
func (linuxOS) Lock() error {
	uid := strconv.Itoa(os.Getuid())

	out, err := exec.Command("loginctl", "show-user", uid, "--property=Display", "--value").Output()
	if err != nil {
		return fmt.Errorf("linux: loginctl show-user: %w", err)
	}

	session := strings.TrimSpace(string(out))
	if session == "" {
		return fmt.Errorf("linux: no graphical session for uid %s", uid)
	}

	return run("loginctl", "lock-session", session)
}

func run(name string, args ...string) error {
	if out, err := exec.Command(name, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("linux: %s: %w: %s", name, err, string(out))
	}
	return nil
}
