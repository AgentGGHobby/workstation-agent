package mqtt

import (
	"fmt"
	"log"
	"sync"
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

	subscribeQoS = 1
)

type Config struct {
	Broker   string
	ClientID string
	Username string
	Password string
}

// Handler verarbeitet eine eingehende Nachricht.
type Handler func(topic string, payload []byte, retained bool)

type Client struct {
	client paho.Client
	topic  string

	mu   sync.Mutex
	subs map[string]Handler
}

// StateTopic liefert das State-Topic des jeweiligen OS, z. B. tower/state/windows.
func StateTopic(osID string) string {
	return "tower/state/" + osID
}

// Connect verbindet sich mit Retry, Reconnect und Last Will.
// Nach jedem (Re-)Connect wird "online" retained auf das State-Topic des OS gesendet
// und alle registrierten Subscriptions werden erneut abonniert.
// Ist der Broker beim Start noch nicht erreichbar, kehrt die Funktion nach connectWait
// ohne Fehler zurück, die Verbindung wird im Hintergrund weiter versucht.
func Connect(cfg Config, osID string) (*Client, error) {
	if osID == "" {
		return nil, fmt.Errorf("mqtt: os id must not be empty")
	}
	topic := StateTopic(osID)

	c := &Client{
		topic: topic,
		subs:  make(map[string]Handler),
	}

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

	opts.SetOnConnectHandler(func(cl paho.Client) {
		token := cl.Publish(topic, 1, true, stateOnline)
		if !token.WaitTimeout(publishWait) {
			log.Printf("mqtt: publish %s=%s timed out", topic, stateOnline)
		} else if err := token.Error(); err != nil {
			log.Printf("mqtt: publish %s=%s failed: %v", topic, stateOnline, err)
		}

		c.resubscribe(cl)
	})

	opts.SetConnectionLostHandler(func(_ paho.Client, err error) {
		log.Printf("mqtt: connection lost: %v", err)
	})

	c.client = paho.NewClient(opts)

	token := c.client.Connect()
	if token.WaitTimeout(connectWait) {
		if err := token.Error(); err != nil {
			return nil, fmt.Errorf("mqtt connect: %w", err)
		}
	} else {
		log.Printf("mqtt: broker not reachable yet, retrying in background")
	}

	return c, nil
}

// Subscribe registriert einen Handler für ein Topic. Die Subscription wird bei
// jedem Reconnect automatisch erneuert. Ist die Verbindung noch nicht offen,
// wird sie beim Verbinden nachgeholt.
func (c *Client) Subscribe(topic string, h Handler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.subs[topic] = h

	if !c.client.IsConnectionOpen() {
		return nil
	}
	return c.subscribe(c.client, topic, h)
}

func (c *Client) resubscribe(cl paho.Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for topic, h := range c.subs {
		if err := c.subscribe(cl, topic, h); err != nil {
			log.Printf("mqtt: %v", err)
		}
	}
}

func (c *Client) subscribe(cl paho.Client, topic string, h Handler) error {
	token := cl.Subscribe(topic, subscribeQoS, func(_ paho.Client, m paho.Message) {
		h(m.Topic(), m.Payload(), m.Retained())
	})
	if !token.WaitTimeout(publishWait) {
		return fmt.Errorf("subscribe %s timed out", topic)
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("subscribe %s: %w", topic, err)
	}
	return nil
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
