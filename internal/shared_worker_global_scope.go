//go:build js && wasm

package internal

import (
	"context"
	"fmt"
	"github.com/hack-pad/safejs"
)

type SharedWorkerGlobalScope struct {
	self safejs.Value
}

func WrapSharedWorkerGlobalScope(v safejs.Value) (*SharedWorkerGlobalScope, error) {
	someMethod, err := v.Get("onconnect")
	if err != nil {
		return nil, err
	}
	if truthy, err := someMethod.Truthy(); err != nil || !truthy {
		return nil, fmt.Errorf("invalid SharedWorkerGlobalScope value: onconnect is not a function")
	}
	return &SharedWorkerGlobalScope{v}, nil
}

func (p *SharedWorkerGlobalScope) Listen(ctx context.Context) (_ <-chan ConnectEvent, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	events := make(chan ConnectEvent)
	connectHandler, err := nonBlocking(func(args []safejs.Value) {
		events <- parseConnectEvent(args[0])
	})
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		_, err := p.self.Call("removeEventListener", "connect", connectHandler)
		if err == nil {
			connectHandler.Release()
		}
		close(events)
	}()
	_, err = p.self.Call("addEventListener", "connect", connectHandler)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (p *SharedWorkerGlobalScope) Close() error {
	_, err := p.self.Call("close")
	return err
}
