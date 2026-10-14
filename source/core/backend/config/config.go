package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Broker   string
	User     string
	Password string
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
