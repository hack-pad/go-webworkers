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

func (p *MessagePort) Listen(ctx context.Context) (_ <-chan MessageEvent, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	events := make(chan MessageEvent)
	messageHandler, err := nonBlocking(func(args []safejs.Value) {
		events <- parseMessageEvent(args[0])
	})
	if err != nil {
		return nil, err
	}
	errorHandler, err := nonBlocking(func(args []safejs.Value) {
		events <- parseMessageEvent(args[0])
	})
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		_, err := p.jsMessagePort.Call("removeEventListener", "message", messageHandler)
		if err == nil {
			messageHandler.Release()
		}
		_, err = p.jsMessagePort.Call("removeEventListener", "messageerror", errorHandler)
		if err == nil {
			errorHandler.Release()
		}
		close(events)
	}()
	_, err = p.jsMessagePort.Call("addEventListener", "message", messageHandler)
	if err != nil {
		return nil, err
	}
	_, err = p.jsMessagePort.Call("addEventListener", "messageerror", errorHandler)
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

func (p *MessagePort) Close() error {
	_, err := p.jsMessagePort.Call("close")
	return err
}
