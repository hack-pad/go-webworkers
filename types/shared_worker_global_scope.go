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

// Name returns the name that the Worker was (optionally) given when it was created.
func (p *SharedWorkerGlobalScope) Name() (string, error) {
	v, err := p.self.Get("name")
	if err != nil {
		return "", err
	}
	return v.String()
}

// Location returns the WorkerLocation in the form of url.URL for this worker.
func (p *SharedWorkerGlobalScope) Location() (*WorkerLocation, error) {
	loc, err := p.self.Get("location")
	if err != nil {
		return nil, err
	}

	location := &WorkerLocation{}
	l := []struct {
		target *string
		prop   string
	}{
		{&location.Hash, "hash"},
		{&location.Host, "host"},
		{&location.HostName, "hostname"},
		{&location.Href, "href"},
		{&location.Origin, "origin"},
		{&location.PathName, "pathname"},
		{&location.Port, "port"},
		{&location.Protocol, "protocol"},
		{&location.Search, "search"},
	}

	for _, entry := range l {
		v, err := loc.Get(entry.prop)
		if err != nil {
			return nil, err
		}
		vv, err := v.String()
		if err != nil {
			return nil, err
		}
		*entry.target = vv
	}

	return location, nil
}
