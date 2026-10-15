//go:build linux

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

const envPath = "/etc/workstation-agent/.env"

func main() {
	osID, err := detectOSID()
	if err != nil {
		log.Fatalf("%v", err)
	}
	clientID := "workstation-agent-" + osID

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

	dispatcher := commands.New(linuxOS{})
	if err := client.Subscribe(commands.Topic, dispatcher.Handle); err != nil {
		log.Printf("subscribe %s: %v", commands.Topic, err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Printf("shutting down")
	client.Close()
}
