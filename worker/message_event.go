//go:build js && wasm
// +build js,wasm

package worker

import (
	"github.com/hack-pad/safejs"
	"github.com/pkg/errors"
)

type MessageEvent struct {
	data   safejs.Value
	err    error
	target *messagePort
}

func (e MessageEvent) Data() (safejs.Value, error) {
	return e.data, errors.Wrapf(e.err, "failed to parse MessageEvent %+v", e.data)
}

func parseMessageEvent(v safejs.Value) MessageEvent {
	value, err := v.Get("target")
	if err != nil {
		return MessageEvent{err: err}
	}
	target, err := wrapMessagePort(value)
	if err != nil {
		return MessageEvent{err: err}
	}
	data, err := v.Get("data")
	if err != nil {
		return MessageEvent{err: err}
	}
	return MessageEvent{
		data:   data,
		target: target,
	}
}
