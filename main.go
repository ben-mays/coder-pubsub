package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/ben-mays/coder-pubsub/client"
	"github.com/ben-mays/coder-pubsub/pubsub"
)

// Example server using the pubsub lib with a HTTP transport.
func main() {
	port := "3939"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	pubsub := pubsub.NewPubSub()

	buffers := sync.Pool{
		New: func() interface{} {
			// Intentionally small to test batching.
			return make([]byte, 16)
		},
	}

	http.HandleFunc("/publish", func(writer http.ResponseWriter, req *http.Request) {
		// The buffer slice is shared across goroutines and we must be careful not to modify it,
		// to avoid GC pressure we use a pool of buffers (definitely overkill).
		for {
			buf := buffers.Get().([]byte)
			n, err := req.Body.Read(buf)
			if n != 0 {
				pubsub.Publish(buf[0:n])
			}
			if err == io.EOF {
				return
			}
		}
	})

	http.HandleFunc("/subscribe", func(rw http.ResponseWriter, req *http.Request) {
		client := &client.Client{
			// In this example, `In` is not used since we expose a separate HTTP endpoint for broadcasting.
			In:  make(chan []byte),
			Out: make(chan []byte),
		}
		pubsub.Register(req.RemoteAddr, client)
		// Unregister here to avoid leaking the writer. We don't want to risk closing the channel directly since other
		// requests may be broadcasting to the chan. Instead we just remove it from the registry and let the runtime GC it.
		defer pubsub.Unregister(req.RemoteAddr)
		for {
			select {
			case msg, ok := <-client.Out:
				if !ok {
					return
				}
				_, err := rw.Write(msg)
				if err != nil {
					return
				}
				// Force internal RW buffer to flush
				if f, ok := rw.(http.Flusher); ok {
					f.Flush()
				}
				// Place buffer back into the pool
				buffers.Put(msg)
			}
		}
	})

	pubsub.Start()
	defer pubsub.Stop()

	fmt.Printf("running PubSub on %s\n", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), http.DefaultServeMux)
}
