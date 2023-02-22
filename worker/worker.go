//go:build js && wasm
// +build js,wasm

// Package worker provides a Web Workers driver for Go code compiled to WebAssembly.
package worker

import "errors"

// Worker is a Web Worker, which represents a background task that can be created via script.
// Workers can send messages back to its creator.
type Worker struct{}

// NewWorker returns a new Worker
func NewWorker() (*Worker, error) {
	return nil, errors.New("not implemented")
}
