package pubsub_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/ben-mays/coder-pubsub/client"
	"github.com/ben-mays/coder-pubsub/pubsub"
	"github.com/stretchr/testify/assert"
)

// Wrapper around a client that captures all outbound messages.
type statefulClient struct {
	c    *client.Client
	msgs []string
}

func newClient(wg *sync.WaitGroup) *statefulClient {
	sc := &statefulClient{
		c:    &client.Client{In: make(chan []byte), Out: make(chan []byte)},
		msgs: make([]string, 0),
	}
	// should be safe as the only reader
	go func() {
		for {
			select {
			case msg, ok := <-sc.c.Out:
				if !ok {
					return
				}
				sc.msgs = append(sc.msgs, string(msg))
				wg.Done()
			}
		}
	}()
	return sc
}

func TestPubSub(t *testing.T) {
	tests := map[string]struct {
		subs int
		msgs []string
	}{
		"no subs, no panics": {
			subs: 0,
			msgs: []string{"hello", "world"},
		},
		"normal case": {
			subs: 5,
			msgs: []string{"hello", "world"},
		},
		"1000s of subs": {
			subs: 1000,
			msgs: []string{"hello", "world"},
		},
		"1000s of messags": {
			subs: 50,
			// Generates 10k random length ~ASCII strings
			msgs: func() []string {
				res := make([]string, 10000)
				for i := 0; i < 10000; i++ {
					mlen := rand.Intn(1000)
					str := make([]byte, mlen)
					for j := 0; j < mlen; j++ {
						str[j] = byte(rand.Intn(256))
					}
					res[i] = string(str)
				}
				return res
			}(),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ps := pubsub.NewPubSub()
			ps.Start()
			defer ps.Stop()
			clients := make([]*statefulClient, tt.subs)
			wg := sync.WaitGroup{}
			wg.Add(len(tt.msgs) * tt.subs)
			for i := 0; i < tt.subs; i++ {
				c := newClient(&wg)
				ps.Register(fmt.Sprintf("%d", i), c.c)
				clients[i] = c
			}
			for _, m := range tt.msgs {
				ps.Publish([]byte(m))
			}
			wg.Wait()
			for i := 0; i < tt.subs; i++ {
				assert.Equal(t, tt.msgs, clients[i].msgs)
			}
		})
	}
}
