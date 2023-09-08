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

// Listen listens on the "connect" events, until the ctx is canceled.
func (p *SharedWorkerGlobalScope) Listen(ctx context.Context) (<-chan MessageEvent, error) {
	events, err := listen(ctx, p.self, "connect")
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
