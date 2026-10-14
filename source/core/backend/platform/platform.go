package platform

import "errors"

// ErrNotImplemented wird von allen Befehlen der Stub-Implementierung zurückgegeben,
// solange die OS-spezifische Implementierung noch fehlt.
var ErrNotImplemented = errors.New("platform: not implemented")

// OS beschreibt die Befehle, die der Agent auf dem Rechner ausführen kann.
type OS interface {
	Shutdown() error
	Reboot() error
	Sleep() error
	Lock() error
}

// Stub erfüllt das Interface, tut aber nichts und meldet bei jedem Befehl
// ErrNotImplemented. Er dient als Platzhalter, bis Windows und Linux umgesetzt sind.
type Stub struct{}

var _ OS = Stub{}

func (Stub) Shutdown() error { return ErrNotImplemented }
func (Stub) Reboot() error   { return ErrNotImplemented }
func (Stub) Sleep() error    { return ErrNotImplemented }
func (Stub) Lock() error     { return ErrNotImplemented }