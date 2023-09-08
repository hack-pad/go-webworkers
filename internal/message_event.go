//go:build js && wasm

package internal

import (
	"github.com/hack-pad/safejs"
	"github.com/pkg/errors"
)

// MessageEvent is received from the channel returned by Listen().
// Represents a JS MessageEvent.
type MessageEvent struct {
	data   safejs.Value
	err    error
	target *MessagePort
	ports  []*MessagePort
}

// Data returns this event's data or a parse error
func (e MessageEvent) Data() (safejs.Value, error) {
	return e.data, errors.Wrapf(e.err, "failed to parse MessageEvent %+v", e.data)
}

// Ports returns this event's ports or a parse error
func (e MessageEvent) Ports() ([]*MessagePort, error) {
	return e.ports, errors.Wrapf(e.err, "failed to parse MessageEvent %+v", e.data)
}

func parseMessageEvent(v safejs.Value) MessageEvent {
	value, err := v.Get("target")
	if err != nil {
		return MessageEvent{err: err}
	}
	target, err := WrapMessagePort(value)
	if err != nil {
		return MessageEvent{err: err}
	}
	data, err := v.Get("data")
	if err != nil {
		return MessageEvent{err: err}
	}
	ports, err := v.Get("ports")
	if err != nil {
		return MessageEvent{err: err}
	}
	portsLen, err := ports.Length()
	if err != nil {
		return MessageEvent{err: err}
	}
	var msgports []*MessagePort
	for i := 0; i < portsLen; i++ {
		port, err := ports.Index(i)
		if err != nil {
			return MessageEvent{err: err}
		}
		msgport, err := WrapMessagePort(port)
		if err != nil {
			return MessageEvent{err: err}
		}
		msgports = append(msgports, msgport)
	}
	return MessageEvent{
		data:   data,
		target: target,
		ports:  msgports,
	}
}
