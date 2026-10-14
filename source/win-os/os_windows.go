package main

import (
	"fmt"
	"os/exec"
	"syscall"

	"github.com/AgentGGHobby/workstation-agent/source/core/backend/platform"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	powrprof = syscall.NewLazyDLL("powrprof.dll")

	procLockWorkStation = user32.NewProc("LockWorkStation")
	procSetSuspendState = powrprof.NewProc("SetSuspendState")
)

type windowsOS struct{}

var _ platform.OS = windowsOS{}

func (windowsOS) Shutdown() error {
	return run("shutdown", "/s", "/t", "0")
}

func (windowsOS) Reboot() error {
	return run("shutdown", "/r", "/t", "0")
}

// Sleep ruft SetSuspendState(hibernate=false, force=false, wakeupEventsDisabled=false).
// Die Funktion kehrt erst nach dem Aufwachen zurück.
func (windowsOS) Sleep() error {
	r1, _, err := procSetSuspendState.Call(0, 0, 0)
	if byte(r1) == 0 {
		return fmt.Errorf("windows: SetSuspendState: %w", err)
	}
	return nil
}

func (windowsOS) Lock() error {
	r1, _, err := procLockWorkStation.Call()
	if uint32(r1) == 0 {
		return fmt.Errorf("windows: LockWorkStation: %w", err)
	}
	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("windows: %s: %w: %s", name, err, string(out))
	}
	return nil
}
