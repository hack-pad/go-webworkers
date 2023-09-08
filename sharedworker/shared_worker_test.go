//go:build js && wasm

package sharedworker

import (
	"context"
	"fmt"
	"testing"

	"github.com/hack-pad/safejs"
)

var (
	jsJSON       = safejs.MustGetGlobal("JSON")
	jsUint8Array = safejs.MustGetGlobal("Uint8Array")
)

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

func TestNew(t *testing.T) {
	t.Parallel()
	const messageText = "Hello, world!"
	blobURL := makeBlobURL(t, []byte(fmt.Sprintf(`"use strict";
onconnect = (e) => {
    const port = e.ports[0];
	port.postMessage(self.name + ": " + %q);
};
`, messageText)), "text/javascript")
	workerName := "worker"
	worker, err := New(blobURL, workerName)
	if err != nil {
		t.Fatal(err)
	}

	if worker.URL() != blobURL {
		t.Fatalf("url expect=%q, got=%q", blobURL, worker.URL())
	}

	if worker.Name() != workerName {
		t.Fatalf("url expect=%q, got=%q", workerName, worker.Name())
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
		t.Fatal(err)
	}
	dataStr, err := data.String()
	if err != nil {
		t.Fatal(err)
	}
	if msg := workerName + ": " + messageText; dataStr != msg {
		t.Errorf("Expected %q, got %q", msg, dataStr)
	}
}

func TestNewFromScript(t *testing.T) {
	t.Parallel()
	const messageText = "Hello, world!"
	script := fmt.Sprintf(`
"use strict";

onconnect = (e) => {
    const port = e.ports[0];
	port.postMessage(self.name + ": " + %q);
};
`, messageText)
	workerName := "worker"
	worker, err := NewFromScript(script, workerName)
	if err != nil {
		t.Fatal(err)
	}
	if worker.URL() == "" {
		t.Fatal("url unexpect to be empty")
	}

	if worker.Name() != workerName {
		t.Fatalf("url expect=%q, got=%q", workerName, worker.Name())
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
		t.Fatal(err)
	}
	dataStr, err := data.String()
	if err != nil {
		t.Fatal(err)
	}
	if msg := workerName + ": " + messageText; dataStr != msg {
		t.Errorf("Expected %q, got %q", msg, dataStr)
	}
}

func TestWorkerPostMessage(t *testing.T) {
	t.Parallel()
	const pingPongScript = `
"use strict";

onconnect = (e) => {
    const port = e.ports[0];
	port.onmessage = (event) => {
		port.postMessage(event.data + " pong!");
	};
};
`
	pingMessage, err := safejs.ValueOf("ping!")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("listen before post", func(t *testing.T) {
		t.Parallel()
		worker, err := NewFromScript(pingPongScript, "")
		if err != nil {
			t.Fatal(err)
		}

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
}
