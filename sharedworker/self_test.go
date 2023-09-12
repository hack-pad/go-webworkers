//go:build js && wasm

package sharedworker

import (
	"testing"
)

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

func TestSelfLocation(t *testing.T) {
	t.Skip("This test case only runs inside a worker")
	t.Parallel()
	self, err := Self()
	if err != nil {
		t.Fatal(err)
	}
	loc, err := self.Location()
	if err != nil {
		t.Fatal(err)
	}
	if loc.String() == "" {
		t.Errorf("Expected %q, got %q", loc.String(), "")
	}
}
