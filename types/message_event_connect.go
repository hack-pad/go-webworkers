package types

import (
	"github.com/hack-pad/safejs"
	"github.com/pkg/errors"
)

// MessageEventConnect represents a JS MessageEvent received from the "connect" event.
type MessageEventConnect struct {
	ports []*MessagePort
	err   error
}

// Ports returns this event's ports or a parse error
func (e MessageEventConnect) Ports() ([]*MessagePort, error) {
	return e.ports, errors.Wrapf(e.err, "failed to parse MessageEventConnect %+v", e.ports)
}

func parseMessageEventConnect(v safejs.Value) MessageEventConnect {
	ports, err := v.Get("ports")
	if err != nil {
		return MessageEventConnect{err: err}
	}
	portsLen, err := ports.Length()
	if err != nil {
		return MessageEventConnect{err: err}
	}
	var msgports []*MessagePort
	for i := 0; i < portsLen; i++ {
		port, err := ports.Index(i)
		if err != nil {
			return MessageEventConnect{err: err}
		}
		msgport, err := WrapMessagePort(port)
		if err != nil {
			return MessageEventConnect{err: err}
		}
		msgports = append(msgports, msgport)
	}
	return MessageEventConnect{
		ports: msgports,
	}
}
