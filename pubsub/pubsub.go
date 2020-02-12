package pubsub

import (
	"fmt"
	"sync"

	"github.com/ben-mays/coder-pubsub/client"
)

// PubSub is the central broker for receiving and dispatching messages. It is transport agnostic. PubSub *MUST NOT* close client channels; rather
// a closed channel indicates that the PubSub broker should remove the client from the registry to avoid writing to a closed channel.
type PubSub struct {
	sync.RWMutex
	registry map[string]*client.Client
	running  bool
	cancel   chan struct{}
}

// NewPubSub returns an instantiated PubSub broker.
func NewPubSub() *PubSub {
	return &PubSub{
		registry: make(map[string]*client.Client),
		running:  false,
		cancel:   make(chan struct{}),
	}
}

// Register wires up a client to the PubSub broker using the given key. If the broker is running, clients will immediately begin receiving
// messages. If not running, clients will receive messages once Start() is invoked.
func (ps *PubSub) Register(key string, c *client.Client) {
	ps.Lock()
	defer ps.Unlock()

	ps.registry[key] = c

	if ps.running {
		go ps.dispatcher(key, c)
	}

	fmt.Printf("registered client=(%s)\n", key)
}

// Unregister removes the client specified by the given key from the broker.
func (ps *PubSub) Unregister(key string) {
	ps.Lock()
	defer ps.Unlock()

	delete(ps.registry, key)
	fmt.Printf("unregistered client=(%s)\n", key)
}

// Publish broadcasts the given message to all clients.
func (ps *PubSub) Publish(message []byte) {
	ps.RLock()
	defer ps.RUnlock()
	for _, v := range ps.registry {
		v.Out <- message
	}
}

// Lifecycle methods

// Start will start processing messages from all clients.
func (ps *PubSub) Start() {
	ps.Lock()
	defer ps.Unlock()
	ps.running = true

	for k, v := range ps.registry {
		go ps.dispatcher(k, v)
	}
}

// Stop will stop all client processing _but_ will not affect client registration. A stopped broker can be restarted.
func (ps *PubSub) Stop() {
	ps.Lock()
	defer ps.Unlock()
	ps.running = false
	// Take the lock and send len(registry) cancel messages
	for i := 0; i < len(ps.registry); i++ {
		ps.cancel <- struct{}{}
	}
}

// Running returns true if the broker is actively processing messages.
func (ps *PubSub) Running() bool {
	return ps.running
}

func (ps *PubSub) dispatcher(key string, c *client.Client) {
	for {
		select {
		case <-ps.cancel:
			return
		case msg, ok := <-c.In:
			if !ok {
				fmt.Printf("error: channed closed prematurely client=(%s)\n", key)
				return
			}
			ps.Publish(msg)
		}
	}
}
