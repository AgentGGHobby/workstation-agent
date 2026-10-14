package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Broker   string
	User     string
	Password string
}

var brokerPattern = regexp.MustCompile(`^(tcp|ssl)://[^\s/]+:[0-9]+$`)

// ValidateBroker prüft, ob der Broker im Format tcp://HOST:PORT oder ssl://HOST:PORT vorliegt.
func ValidateBroker(broker string) error {
	if !brokerPattern.MatchString(broker) {
		return errors.New("broker must look like tcp://HOST:PORT")
	}
	return nil
}

func validateValue(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s must not be empty", name)
	}
	if strings.ContainsAny(value, "'\r\n") {
		return fmt.Errorf("%s must not contain ', CR or LF", name)
	}
	return nil
}

// Validate prüft alle Felder, bevor sie in eine .env geschrieben werden.
func (c Config) Validate() error {
	if err := ValidateBroker(c.Broker); err != nil {
		return err
	}
	if err := validateValue("user", c.User); err != nil {
		return err
	}
	return validateValue("password", c.Password)
}

func Load(path string) (Config, error) {
	env, err := godotenv.Read(path)
	if err != nil {
		return Config{}, fmt.Errorf("config: read %s: %w", path, err)
	}

	cfg := Config{
		Broker:   env["MQTT_BROKER"],
		User:     env["MQTT_USER"],
		Password: env["MQTT_PASSWORD"],
	}

	var missing []string
	if cfg.Broker == "" {
		missing = append(missing, "MQTT_BROKER")
	}
	if cfg.User == "" {
		missing = append(missing, "MQTT_USER")
	}
	if cfg.Password == "" {
		missing = append(missing, "MQTT_PASSWORD")
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("config: missing variables in %s: %s", path, strings.Join(missing, ", "))
	}

	return cfg, nil
}

// Write prüft die Werte und schreibt sie als .env. Die Werte stehen in einfachen
// Anführungszeichen, damit Sonderzeichen im Passwort unverändert bleiben.
func Write(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("config: create dir: %w", err)
	}

	content := fmt.Sprintf("MQTT_BROKER='%s'\nMQTT_USER='%s'\nMQTT_PASSWORD='%s'\n", cfg.Broker, cfg.User, cfg.Password)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("config: write %s: %w", path, err)
	}

	return nil
}
