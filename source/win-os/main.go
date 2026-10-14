//go:build windows

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/AgentGGHobby/workstation-agent/source/core/backend"
	"github.com/AgentGGHobby/workstation-agent/source/core/backend/commands"
	"github.com/AgentGGHobby/workstation-agent/source/core/backend/config"
	"github.com/AgentGGHobby/workstation-agent/source/core/backend/mqtt"
)

const (
	osID     = "windows"
	clientID = "workstation-agent-windows"
	envPath  = `C:\ProgramData\workstation-agent\.env`
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "configure" {
		os.Exit(runConfigure(os.Args[2:]))
	}

	log.Printf("workstation-agent %s (%s)", backend.Version, osID)

	cfg, err := config.Load(envPath)
	if err != nil {
		log.Fatalf("%v", err)
	}

	client, err := mqtt.Connect(mqtt.Config{
		Broker:   cfg.Broker,
		ClientID: clientID,
		Username: cfg.User,
		Password: cfg.Password,
	}, osID)
	if err != nil {
		log.Fatalf("%v", err)
	}

	dispatcher := commands.New(windowsOS{})
	if err := client.Subscribe(commands.Topic, dispatcher.Handle); err != nil {
		log.Printf("subscribe %s: %v", commands.Topic, err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Printf("shutting down")
	client.Close()
}
