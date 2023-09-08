//go:build js && wasm

package types

import (
	"github.com/hack-pad/safejs"
	"github.com/pkg/errors"
)

// MessageEventMessage represents a JS MessageEvent received from the "message" event.
type MessageEventMessage struct {
	data   safejs.Value
	err    error
	target *MessagePort
}

// Data returns this event's data or a parse error
func (e MessageEventMessage) Data() (safejs.Value, error) {
	return e.data, errors.Wrapf(e.err, "failed to parse MessageEventMessage %+v", e.data)
}

func parseMessageEventMessage(v safejs.Value) MessageEventMessage {
	value, err := v.Get("target")
	if err != nil {
		return MessageEventMessage{err: err}
	}
	target, err := WrapMessagePort(value)
	if err != nil {
		return MessageEventMessage{err: err}
	}
	data, err := v.Get("data")
	if err != nil {
		return MessageEventMessage{err: err}
	}
	return MessageEventMessage{
		data:   data,
		target: target,
	}
}
