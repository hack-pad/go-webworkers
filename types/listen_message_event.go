package types

import (
	"context"

	"github.com/hack-pad/safejs"
)

// listen adds the EventListener on the listener for the specified events.
// It returns a channel, which will send the MessageEvent(s) listened on, until the ctx is canceled.
func listen(ctx context.Context, listener safejs.Value, events ...string) (_ <-chan MessageEvent, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	eventsCh := make(chan MessageEvent)

	var handlers []safejs.Func
	for range events {
		handler, err := nonBlocking(func(args []safejs.Value) {
			eventsCh <- parseMessageEvent(args[0])
		})
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, handler)
	}

	go func() {
		<-ctx.Done()
		for i := range events {
			event, handler := events[i], handlers[i]
			_, err := listener.Call("removeEventListener", event, handler)
			if err == nil {
				handler.Release()
			}
		}
		close(eventsCh)
	}()

	for i := range events {
		event, handler := events[i], handlers[i]
		_, err = listener.Call("addEventListener", event, handler)
		if err != nil {
			return nil, err
		}
	}

	return eventsCh, nil
}

func nonBlocking(fn func(args []safejs.Value)) (safejs.Func, error) {
	return safejs.FuncOf(func(_ safejs.Value, args []safejs.Value) any {
		go fn(args)
		return nil
	})
}
