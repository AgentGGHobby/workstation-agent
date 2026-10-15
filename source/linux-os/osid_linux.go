package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

const osReleasePath = "/etc/os-release"

// osIDAliases bildet die ID aus /etc/os-release auf die OS-ID im Topic ab.
// Alles, was hier nicht steht, wird unverändert übernommen.
var osIDAliases = map[string]string{
	"endeavouros": "endeavour",
	"arch":        "lab",
}

// validOSID erlaubt nur Zeichen, die in einem MQTT-Topic-Segment unbedenklich sind.
var validOSID = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// detectOSID liest die ID des laufenden Systems und gibt die OS-ID für das Topic zurück.
func detectOSID() (string, error) {
	f, err := os.Open(osReleasePath)
	if err != nil {
		return "", fmt.Errorf("os-release: %w", err)
	}
	defer f.Close()

	return parseOSRelease(f)
}

// parseOSRelease sucht die Zeile ID=... und wendet die Abbildung an.
func parseOSRelease(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		value, found := strings.CutPrefix(line, "ID=")
		if !found {
			continue
		}

		id := strings.ToLower(strings.Trim(strings.TrimSpace(value), `"'`))
		if alias, ok := osIDAliases[id]; ok {
			id = alias
		}

		if !validOSID.MatchString(id) {
			return "", fmt.Errorf("os-release: ungültige ID %q", id)
		}

		return id, nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("os-release: %w", err)
	}

	return "", errors.New("os-release: keine ID-Zeile gefunden")
}
