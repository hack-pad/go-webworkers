//go:build js && wasm
// +build js,wasm

package worker

import "testing"

func TestNewWorker(t *testing.T) {
	t.Parallel()
	_, err := NewWorker()
	if err == nil || err.Error() != "not implemented" {
		t.Error("NewWorker should not be implemented, got:", err)
	}
}
