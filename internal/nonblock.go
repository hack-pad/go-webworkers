//go:build js && wasm

package internal

import "github.com/hack-pad/safejs"

func nonBlocking(fn func(args []safejs.Value)) (safejs.Func, error) {
	return safejs.FuncOf(func(_ safejs.Value, args []safejs.Value) any {
		go fn(args)
		return nil
	})
}
