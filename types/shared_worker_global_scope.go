//go:build js && wasm

package types

import (
	"context"
	"fmt"

	"github.com/hack-pad/safejs"
)

type SharedWorkerGlobalScope struct {
	self safejs.Value
}

func WrapSharedWorkerGlobalScope(v safejs.Value) (*SharedWorkerGlobalScope, error) {
	someMethod, err := v.Get("SharedWorkerGlobalScope")
	if err != nil {
		return nil, err
	}
	if truthy, err := someMethod.Truthy(); err != nil || !truthy {
		return nil, fmt.Errorf("invalid SharedWorkerGlobalScope value: SharedWorkerGlobalScope is not a function")
	}
	return &SharedWorkerGlobalScope{v}, nil
}

// Listen listens on the "connect" events, until the ctx is canceled.
func (p *SharedWorkerGlobalScope) Listen(ctx context.Context) (<-chan MessageEventConnect, error) {
	events, err := listen(ctx, p.self, parseMessageEventConnect, "connect")
	if err != nil {
		return nil, err
	}
	return events, nil
}

// Close discards any tasks queued in the global scope's event loop, effectively closing this particular scope.
func (p *SharedWorkerGlobalScope) Close() error {
	_, err := p.self.Call("close")
	return err
}
