//go:build js && wasm

// Package sharedworker provides a Shared Web Workers driver for Go code compiled to WebAssembly.
package sharedworker

import (
	"context"

	"github.com/hack-pad/go-webworkers/types"

	"github.com/hack-pad/safejs"
)

var (
	jsWorker = safejs.MustGetGlobal("SharedWorker")
	jsURL    = safejs.MustGetGlobal("URL")
	jsBlob   = safejs.MustGetGlobal("Blob")
)

// SharedWorker is a Shared Web Worker, which represents a background task created via a script.
// Use Listen() and PostMessage() to communicate with the worker.
type SharedWorker struct {
	url     string
	name    string
	worker  safejs.Value
	msgport *types.MessagePort
}

// New starts a worker with the given script's URL and name
func New(url, name string) (*SharedWorker, error) {
	worker, err := jsWorker.New(url, name)
	if err != nil {
		return nil, err
	}
	port, err := worker.Get("port")
	if err != nil {
		return nil, err
	}
	msgport, err := types.WrapMessagePort(port)
	if err != nil {
		return nil, err
	}
	return &SharedWorker{
		url:     url,
		name:    name,
		msgport: msgport,
		worker:  worker,
	}, nil
}

// NewFromScript is like New, but starts the worker with the given script (in JavaScript)
func NewFromScript(jsScript, name string) (*SharedWorker, error) {
	blob, err := jsBlob.New([]any{jsScript}, map[string]any{
		"type": "text/javascript",
	})
	if err != nil {
		return nil, err
	}
	objectURL, err := jsURL.Call("createObjectURL", blob)
	if err != nil {
		return nil, err
	}
	objectURLStr, err := objectURL.String()
	if err != nil {
		return nil, err
	}
	return New(objectURLStr, name)
}

// URL returns the script URL of the worker
func (w *SharedWorker) URL() string {
	return w.url
}

// Name returns the name of the worker
func (w *SharedWorker) Name() string {
	return w.name
}

// PostMessage sends data in a message to the worker, optionally transferring ownership of all items in transfers.
//
// The data may be any value handled by the "structured clone algorithm", which includes cyclical references.
//
// Transfers is an optional array of Transferable objects to transfer ownership of.
// If the ownership of an object is transferred, it becomes unusable in the context it was sent from and becomes available only to the worker it was sent to.
// Transferable objects are instances of classes like ArrayBuffer, MessagePort or ImageBitmap objects that can be transferred.
// null is not an acceptable value for transfer.
func (w *SharedWorker) PostMessage(data safejs.Value, transfers []safejs.Value) error {
	return w.msgport.PostMessage(data, transfers)
}

// Listen sends message events on a channel for events fired by self.postMessage() calls inside the Worker's global scope.
// Stops the listener and closes the channel when ctx is canceled.
func (w *SharedWorker) Listen(ctx context.Context) (<-chan types.MessageEvent, error) {
	return w.msgport.Listen(ctx)
}
