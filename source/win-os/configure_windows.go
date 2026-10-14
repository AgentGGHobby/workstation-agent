//go:build windows

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AgentGGHobby/workstation-agent/source/core/backend/config"
)

// runConfigure schreibt die .env aus den Kommandozeilenargumenten.
// Der Rückgabewert ist der Exit-Code des Prozesses, den der Installer auswertet.
func runConfigure(args []string) int {
	fs := flag.NewFlagSet("configure", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	broker := fs.String("broker", "", "MQTT-Broker, z. B. tcp://192.168.178.23:1883")
	user := fs.String("user", "", "MQTT-Benutzer")
	password := fs.String("password", "", "MQTT-Passwort")
	path := fs.String("path", envPath, "Zielpfad der .env")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg := config.Config{
		Broker:   *broker,
		User:     *user,
		Password: *password,
	}

	if err := config.Write(*path, cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
