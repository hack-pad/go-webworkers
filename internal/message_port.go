//go:build js && wasm

package internal

import (
	"context"
	"fmt"

	"github.com/hack-pad/safejs"
)

type MessagePort struct {
	jsMessagePort safejs.Value
}

func WrapMessagePort(v safejs.Value) (*MessagePort, error) {
	someMethod, err := v.Get("postMessage")
	if err != nil {
		return nil, err
	}
	if truthy, err := someMethod.Truthy(); err != nil || !truthy {
		return nil, fmt.Errorf("invalid MessagePort value: postMessage is not a function")
	}
	return &MessagePort{v}, nil
}

func (p *MessagePort) PostMessage(data safejs.Value, transfers []safejs.Value) error {
	args := append([]any{data}, toJSSlice(transfers))
	_, err := p.jsMessagePort.Call("postMessage", args...)
	return err
}

func toJSSlice[Type any](slice []Type) []any {
	newSlice := make([]any, len(slice))
	for i := range slice {
		newSlice[i] = slice[i]
	}
	return newSlice
}

// Listen starts the MessagePort to listen on the "message" and "messageerror" events, until the ctx is canceled.
func (p *MessagePort) Listen(ctx context.Context) (<-chan MessageEvent, error) {
	events, err := listen(ctx, p.jsMessagePort, "message", "messageerror")
	if err != nil {
		return nil, err
	}

	if start, err := p.jsMessagePort.Get("start"); err == nil {
		if truthy, err := start.Truthy(); err == nil && truthy {
			if _, err := p.jsMessagePort.Call("start"); err != nil {
				return nil, err
			}
		}
	}
	return events, nil
}
