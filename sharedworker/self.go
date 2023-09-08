//go:build js && wasm

package sharedworker

import (
	"context"

	"github.com/hack-pad/go-webworkers/types"

	"github.com/hack-pad/safejs"
)

// GlobalSelf represents the global scope, named "self", in the context of using SharedWorkers.
// Supports receiving connection via Listen(), where each of the ConnectEvent has Ports() whose
// first element represents the MessagePort connected with the channel with its parent,
// which in turns support receiving message via its Listen() and PostMessage().
type GlobalSelf struct {
	self  safejs.Value
	scope *types.SharedWorkerGlobalScope
}

// Self returns the global "self"
func Self() (*GlobalSelf, error) {
	self, err := safejs.Global().Get("self")
	if err != nil {
		return nil, err
	}
	scope, err := types.WrapSharedWorkerGlobalScope(self)
	if err != nil {
		return nil, err
	}
	return &GlobalSelf{
		self:  self,
		scope: scope,
	}, nil
}

// Listen sends message events representing the connect event on a channel for events fired
// by connection calls to this worker from within the parent scope.
// Users are expected to call the Ports() on the MessageEvent, and take the 1st one as the target MessagePort.
// Stops the listener and closes the channel when ctx is canceled.
func (s *GlobalSelf) Listen(ctx context.Context) (<-chan types.MessageEventConnect, error) {
	return s.scope.Listen(ctx)
}

// Close discards any tasks queued in the global scope's event loop, effectively closing this particular scope.
func (s *GlobalSelf) Close() error {
	return s.scope.Close()
}

// Name returns the name that the Worker was (optionally) given when it was created.
func (s *GlobalSelf) Name() (string, error) {
	name, err := s.self.Get("name")
	if err != nil {
		return "", err
	}
	return name.String()
}
