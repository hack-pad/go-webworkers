//go:build js && wasm

package sharedworker

import (
	"testing"

	"github.com/hack-pad/safejs"
)

func TestSelf(t *testing.T) {
	t.Skip("This test case only runs inside a worker")
	t.Parallel()
	self, err := Self()
	if err != nil {
		t.Fatal(err)
	}
	if !self.self.Equal(safejs.MustGetGlobal("self")) {
		t.Error("self is not equal to the global self")
	}
}

func TestSelfName(t *testing.T) {
	t.Skip("This test case only runs inside a worker")
	t.Parallel()
	self, err := Self()
	if err != nil {
		t.Fatal(err)
	}
	name, err := self.Name()
	if err != nil {
		t.Fatal(err)
	}
	if name != "" {
		t.Errorf("Expected %q, got %q", "", name)
	}
}
