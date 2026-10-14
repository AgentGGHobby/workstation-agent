# workstation-agent

[![Build Status](https://img.shields.io/github/actions/workflow/status/AgentGGHobby/workstation-agent/build.yml?label=build)](https://github.com/AgentGGHobby/workstation-agent/actions)
[![Version](https://img.shields.io/github/v/release/AgentGGHobby/workstation-agent)](https://github.com/AgentGGHobby/workstation-agent/releases)
[![Status](https://img.shields.io/badge/status-WIP-yellow)]()
[![License](https://img.shields.io/github/license/AgentGGHobby/workstation-agent)](https://github.com/AgentGGHobby/workstation-agent/blob/main/LICENSE)

Plattformübergreifender Go-Agent zur Fernsteuerung eines Dual-Boot-Rechners über MQTT und Home Assistant.

## Features
- PC-Steuerung per MQTT (herunterfahren, sperren, Ruhezustand, neustarten)
- Status- und OS-Meldung an Home Assistant (ein State-Topic pro Betriebssystem, mit Last Will)
- Optionales Szenen-Feature als eigene Binary, die nur bei Bedarf installiert wird
- Native Installer (MSI, Arch-Paket) mit eigener Konfigurationsabfrage

## Installation
Installer werden über GitHub Actions gebaut und als Artifact bereitgestellt (`.msi` für Windows, `.pkg.tar.zst` für Arch-basierte Systeme). Der Installer fragt Broker, Benutzer und Passwort ab und erzeugt die `.env`. Das Szenen-Feature ist im Installer optional.

## Local Development
1. Repository klonen
2. Abhängigkeiten laden: `go mod download`
3. `.env` mit `MQTT_BROKER`, `MQTT_USER` und `MQTT_PASSWORD` am OS-Pfad des Agenten anlegen (es gibt kein `.env.example`, die Datei erzeugt sonst der Installer)
4. Agent starten: `go run ./source/win-os` (Windows) oder `go run ./source/linux-os` (Linux)

## License
MIT
