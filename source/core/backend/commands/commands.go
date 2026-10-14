package commands

import (
	"log"
	"strings"

	"github.com/AgentGGHobby/workstation-agent/source/core/backend/platform"
)

// Topic ist das Topic, auf dem der Agent Befehle von Home Assistant empfängt.
const Topic = "tower/cmd"

const maxLogPayload = 32

type Dispatcher struct {
	handlers map[string]func() error
}

func New(os platform.OS) *Dispatcher {
	return &Dispatcher{
		handlers: map[string]func() error{
			"shutdown": os.Shutdown,
			"reboot":   os.Reboot,
			"sleep":    os.Sleep,
			"lock":     os.Lock,
		},
	}
}

// Handle verarbeitet eine Nachricht von tower/cmd. Retained Nachrichten werden
// ignoriert, sonst würde ein alter Befehl bei jedem Start des Agenten erneut
// ausgeführt. Unbekannte Befehle, auch Szenen-Befehle, werden nur geloggt.
func (d *Dispatcher) Handle(topic string, payload []byte, retained bool) {
	cmd := strings.ToLower(strings.TrimSpace(string(payload)))

	if retained {
		log.Printf("commands: ignoring retained message on %s: %q", topic, truncate(cmd))
		return
	}

	fn, ok := d.handlers[cmd]
	if !ok {
		log.Printf("commands: unknown command on %s: %q", topic, truncate(cmd))
		return
	}

	log.Printf("commands: running %q", cmd)

	// Eigene Goroutine, damit ein blockierender Befehl (Sleep unter Windows kehrt
	// erst nach dem Aufwachen zurück) die MQTT-Verarbeitung nicht aufhält.
	go func() {
		if err := fn(); err != nil {
			log.Printf("commands: %q failed: %v", cmd, err)
		}
	}()
}

func truncate(s string) string {
	if len(s) > maxLogPayload {
		return s[:maxLogPayload] + "..."
	}
	return s
}
