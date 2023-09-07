//go:build js && wasm

package internal

import (
	"github.com/hack-pad/safejs"
	"github.com/pkg/errors"
)

// ConnectEvent is received from the channel returned by sharedworker.GlobalSelf's Listen().
// Represents a JS MessageEvent for the connect event.
type ConnectEvent struct {
	err   error
	ports []*MessagePort
}

// Ports returns this event's ports or a parse error
func (e ConnectEvent) Ports() ([]*MessagePort, error) {
	return e.ports, errors.Wrapf(e.err, "failed to parse ConnectEvent")
}

func parseConnectEvent(v safejs.Value) ConnectEvent {
	ports, err := v.Get("ports")
	if err != nil {
		return ConnectEvent{err: err}
	}
	portsLen, err := ports.Length()
	if err != nil {
		return ConnectEvent{err: err}
	}
	var msgports []*MessagePort
	for i := 0; i < portsLen; i++ {
		port, err := ports.Index(i)
		if err != nil {
			return ConnectEvent{err: err}
		}
		msgport, err := WrapMessagePort(port)
		if err != nil {
			return ConnectEvent{err: err}
		}
		msgports = append(msgports, msgport)
	}
	return ConnectEvent{
		ports: msgports,
	}
}
