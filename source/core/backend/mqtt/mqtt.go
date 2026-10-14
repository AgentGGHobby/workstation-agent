package mqtt

import (
	"fmt"
	"log"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

const (
	stateOnline  = "online"
	stateOffline = "offline"

	keepAlive         = 15 * time.Second
	pingTimeout       = 5 * time.Second
	retryInterval     = 5 * time.Second
	maxReconnectDelay = 30 * time.Second
	connectWait       = 10 * time.Second
	publishWait       = 5 * time.Second
)

type Config struct {
	Broker   string
	ClientID string
	Username string
	Password string
}

type Client struct {
	client paho.Client
	topic  string
}

// StateTopic liefert das State-Topic des jeweiligen OS, z. B. tower/state/windows.
func StateTopic(osID string) string {
	return "tower/state/" + osID
}

// Connect verbindet sich mit Retry, Reconnect und Last Will.
// Nach jedem (Re-)Connect wird "online" retained auf das State-Topic des OS gesendet.
// Ist der Broker beim Start noch nicht erreichbar, kehrt die Funktion nach connectWait
// ohne Fehler zurück, die Verbindung wird im Hintergrund weiter versucht.
func Connect(cfg Config, osID string) (*Client, error) {
	if osID == "" {
		return nil, fmt.Errorf("mqtt: os id must not be empty")
	}
	topic := StateTopic(osID)

	opts := paho.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(cfg.ClientID)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)

	opts.SetKeepAlive(keepAlive)
	opts.SetPingTimeout(pingTimeout)

	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(retryInterval)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(maxReconnectDelay)

	opts.SetWill(topic, stateOffline, 1, true)

	opts.SetOnConnectHandler(func(c paho.Client) {
		token := c.Publish(topic, 1, true, stateOnline)
		if !token.WaitTimeout(publishWait) {
			log.Printf("mqtt: publish %s=%s timed out", topic, stateOnline)
			return
		}
		if err := token.Error(); err != nil {
			log.Printf("mqtt: publish %s=%s failed: %v", topic, stateOnline, err)
		}
	})

	opts.SetConnectionLostHandler(func(_ paho.Client, err error) {
		log.Printf("mqtt: connection lost: %v", err)
	})

	client := paho.NewClient(opts)

	token := client.Connect()
	if token.WaitTimeout(connectWait) {
		if err := token.Error(); err != nil {
			return nil, fmt.Errorf("mqtt connect: %w", err)
		}
	} else {
		log.Printf("mqtt: broker not reachable yet, retrying in background")
	}

	return &Client{client: client, topic: topic}, nil
}

// Close meldet "offline" und trennt sauber. Bei einem sauberen Disconnect löst der
// Broker den Last Will nicht aus, deshalb wird "offline" hier selbst gesendet.
func (c *Client) Close() {
	if c.client.IsConnectionOpen() {
		token := c.client.Publish(c.topic, 1, true, stateOffline)
		token.WaitTimeout(publishWait)
	}
	c.client.Disconnect(250)
}
