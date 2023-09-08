//go:build js && wasm

package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hack-pad/safejs"
)

var (
	jsJSON       = safejs.MustGetGlobal("JSON")
	jsUint8Array = safejs.MustGetGlobal("Uint8Array")
)

func TestWorkerOptionsToJSValue(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		description string
		options     Options
		expect      any
	}{
		{
			description: "no options",
			options:     Options{},
			expect:      map[string]any{},
		},
		{
			description: "name",
			options: Options{
				Name: "foo",
			},
			expect: map[string]any{
				"name": "foo",
			},
		},
	} {
		tc := tc // enable parallel sub-tests
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			value, err := tc.options.toJSValue()
			if err != nil {
				t.Fatal(err)
			}
			expect, err := safejs.ValueOf(tc.expect)
			if err != nil {
				t.Fatal(err)
			}
			expectJSON, actualJSON := stringify(t, expect), stringify(t, value)
			if expectJSON != actualJSON {
				t.Errorf("\nExpected %v\nActual:  %v", expectJSON, actualJSON)
			}
		})
	}
}

func stringify(t *testing.T, obj safejs.Value) string {
	t.Helper()
	json, err := jsJSON.Call("stringify", obj)
	if err != nil {
		t.Fatal(err)
	}
	str, err := json.String()
	if err != nil {
		t.Fatal(err)
	}
	return str
}

func makeBlobURL(t *testing.T, contents []byte, contentType string) string {
	t.Helper()
	jsContents, err := jsUint8Array.New(len(contents))
	if err != nil {
		t.Fatal(err)
	}
	_, err = safejs.CopyBytesToJS(jsContents, contents)
	if err != nil {
		t.Fatal(err)
	}
	blob, err := jsBlob.New([]any{jsContents}, map[string]any{
		"type": contentType,
	})
	if err != nil {
		t.Fatal(err)
	}
	url, err := jsURL.Call("createObjectURL", blob)
	if err != nil {
		t.Fatal(err)
	}
	urlString, err := url.String()
	if err != nil {
		t.Fatal(err)
	}
	return urlString
}

func cleanUpWorker(t *testing.T, worker *Worker) {
	t.Cleanup(func() {
		err := worker.Terminate()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestNew(t *testing.T) {
	t.Parallel()
	const messageText = "Hello, world!"
	blobURL := makeBlobURL(t, []byte(fmt.Sprintf(`"use strict";
self.postMessage(%q);
`, messageText)), "text/javascript")
	worker, err := New(blobURL, Options{})
	if err != nil {
		t.Fatal(err)
	}
	cleanUpWorker(t, worker)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	messages, err := worker.Listen(ctx)
	if err != nil {
		t.Fatal(err)
	}
	message := <-messages
	data, err := message.Data()
	if err != nil {
		t.Fatal(err)
	}
	dataStr, err := data.String()
	if err != nil {
		t.Fatal(err)
	}
	if dataStr != messageText {
		t.Errorf("Expected %q, got %q", messageText, dataStr)
	}
}

func TestNewFromScript(t *testing.T) {
	t.Parallel()
	const messageText = "Hello, world!"
	script := fmt.Sprintf(`
"use strict";

self.postMessage(%q);
`, messageText)
	worker, err := NewFromScript(script, Options{})
	if err != nil {
		t.Fatal(err)
	}
	cleanUpWorker(t, worker)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	messages, err := worker.Listen(ctx)
	if err != nil {
		t.Fatal(err)
	}
	message := <-messages
	data, err := message.Data()
	if err != nil {
		t.Fatal(err)
	}
	dataStr, err := data.String()
	if err != nil {
		t.Fatal(err)
	}
	if dataStr != messageText {
		t.Errorf("Expected %q, got %q", messageText, dataStr)
	}
}

func TestWorkerTerminate(t *testing.T) {
	t.Parallel()
	worker, err := NewFromScript(`
"use strict";

self.postMessage("start");
self.setTimeout(() => self.postMessage("done waiting"), 200);
`, Options{})
	if err != nil {
		t.Fatal(err)
	}
	cleanUpWorker(t, worker)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	messages, err := worker.Listen(ctx)
	if err != nil {
		t.Fatal(err)
	}
	message := <-messages
	data, err := message.Data()
	if err != nil {
		t.Fatal(err)
	}
	dataStr, err := data.String()
	if err != nil {
		t.Error(err)
	}
	if dataStr != "start" {
		t.Fatalf("Expected worker to send 'start', got %s", dataStr)
	}

	err = worker.Terminate()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case message := <-messages:
		t.Errorf("Should not receive the delayed message on a terminated worker, got: %v", message)
	case <-time.After(400 * time.Millisecond):
	}
}

func TestWorkerPostMessage(t *testing.T) {
	t.Parallel()
	const pingPongScript = `
"use strict";

self.addEventListener("message", event => {
	self.postMessage(event.data + " pong!")
});
`
	pingMessage, err := safejs.ValueOf("ping!")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("listen before post", func(t *testing.T) {
		t.Parallel()
		worker, err := NewFromScript(pingPongScript, Options{})
		if err != nil {
			t.Fatal(err)
		}
		cleanUpWorker(t, worker)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		messages, err := worker.Listen(ctx)
		if err != nil {
			t.Fatal(err)
		}

		err = worker.PostMessage(pingMessage, nil)
		if err != nil {
			t.Fatal(err)
		}

		message := <-messages
		data, err := message.Data()
		if err != nil {
			t.Fatal(err)
		}
		dataStr, err := data.String()
		if err != nil {
			t.Error(err)
		}
		expectedResponse := "ping! pong!"
		if dataStr != expectedResponse {
			t.Errorf("Expected response %q, got: %q", expectedResponse, dataStr)
		}
	})

	t.Run("listen after post", func(t *testing.T) {
		t.Parallel()
		worker, err := NewFromScript(pingPongScript, Options{})
		if err != nil {
			t.Fatal(err)
		}
		cleanUpWorker(t, worker)

		err = worker.PostMessage(pingMessage, nil)
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		messages, err := worker.Listen(ctx)
		if err != nil {
			t.Fatal(err)
		}

		message := <-messages
		data, err := message.Data()
		if err != nil {
			t.Error(err)
		}
		dataStr, err := data.String()
		if err != nil {
			t.Error(err)
		}
		expectedResponse := "ping! pong!"
		if dataStr != expectedResponse {
			t.Errorf("Expected response %q, got: %q", expectedResponse, dataStr)
		}
	})
}

func TestWorkerStopListen(t *testing.T) {
	t.Parallel()
	const pingPongScript = `
"use strict";

self.addEventListener("message", event => {
	self.postMessage("foo");
	self.postMessage("bar");
});
`
	worker, err := NewFromScript(pingPongScript, Options{})
	if err != nil {
		t.Fatal(err)
	}
	cleanUpWorker(t, worker)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	_, err = worker.Listen(ctx)
	if err != nil {
		t.Fatal(err)
	}

	msg, err := safejs.ValueOf("start")
	if err != nil {
		t.Fatal(err)
	}

	err = worker.PostMessage(msg, nil)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
}
