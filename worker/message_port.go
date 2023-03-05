//go:build js && wasm
// +build js,wasm

package worker

import (
	"context"
	"fmt"

	"github.com/hack-pad/safejs"
)

var jsMessageChannel = safejs.MustGetGlobal("MessageChannel")

type messagePort struct {
	jsMessagePort safejs.Value
}

func newChannel() (*messagePort, *messagePort, error) {
	channel, err := jsMessageChannel.New()
	if err != nil {
		return nil, nil, err
	}
	port1JSValue, err := channel.Get("port1")
	if err != nil {
		return nil, nil, err
	}
	port1, err := wrapMessagePort(port1JSValue)
	if err != nil {
		return nil, nil, err
	}
	if err != nil {
		return nil, nil, err
	}
	port2JSValue, err := channel.Get("port2")
	if err != nil {
		return nil, nil, err
	}
	port2, err := wrapMessagePort(port2JSValue)
	if err != nil {
		return nil, nil, err
	}
	return port1, port2, nil
}

func wrapMessagePort(v safejs.Value) (*messagePort, error) {
	someMethod, err := v.Get("postMessage")
	if err != nil {
		return nil, err
	}
	if truthy, err := someMethod.Truthy(); err != nil || !truthy {
		return nil, fmt.Errorf("invalid MessagePort value: postMessage is not a function")
	}
	return &messagePort{v}, nil
}

func (p *messagePort) PostMessage(data safejs.Value, transfers []safejs.Value) error {
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

func (p *messagePort) Listen(ctx context.Context) (_ <-chan MessageEvent, err error) {
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

func nonBlocking(fn func(args []safejs.Value)) (safejs.Func, error) {
	return safejs.FuncOf(func(_ safejs.Value, args []safejs.Value) any {
		go fn(args)
		return nil
	})
}

func (p *messagePort) Close() error {
	_, err := p.jsMessagePort.Call("close")
	return err
}

func (p *messagePort) jsValue() safejs.Value {
	if p == nil {
		return safejs.Null()
	}
	return p.jsMessagePort
}