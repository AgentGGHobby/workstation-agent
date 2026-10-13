# workstation-agent

[![Build Status](https://img.shields.io/github/actions/workflow/status/AgentGGHobby/workstation-agent/build.yml?label=build)](https://github.com/AgentGGHobby/workstation-agent/actions)
[![Version](https://img.shields.io/github/v/release/AgentGGHobby/workstation-agent)](https://github.com/AgentGGHobby/workstation-agent/releases)
[![Status](https://img.shields.io/badge/status-WIP-yellow)]()
[![License](https://img.shields.io/github/license/AgentGGHobby/workstation-agent)](https://github.com/AgentGGHobby/workstation-agent/blob/main/LICENSE)

Plattformübergreifender Go-Agent zur Fernsteuerung eines Dual-/Multi-Boot-Rechners über MQTT und Home Assistant.

## Features
- PC-Steuerung per MQTT (herunterfahren, sperren, Ruhezustand, neustarten)
- Status- und OS-Meldung an Home Assistant
- Gemeinsamer Kern für Windows, Linux und ein drittes Betriebssystem
- Native Installer (MSI, Arch-Paket) mit eigener Konfigurationsabfrage

## Installation
Installer werden über GitHub Actions gebaut und als Artifact bereitgestellt (`.msi` für Windows, `.pkg.tar.zst` für Arch-basierte Systeme).

## Local Development
1. Clone the repo
2. Copy `.env.example` to `.env` and fill in real values
3. Install dependencies: `go mod download`
4. Start the agent: `go run ./core/backend`

## License
MIT
