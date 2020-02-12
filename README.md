## Pubsub

PubSub is a transport-agnostic in-memory pubsub library. A bi-directional `Client` can be registered with the pubsub broker, either before starting or during execution. Incoming messages on the `Client.In` channel will be broadcast to all clients. The broker will broadcast to all registered clients on the `Client.Out` channel. The PubSub broker *will never* close a client channel - transport-level implementations must take care to align the transport connection lifecycle with the client.

This simple abstraction was used to allow multiple transports (e.g. http, websocket, tcp) to simply wire up their respective connections to the `Client` channels.

## HTTP Server

A bare-bones HTTP/1.1 implementation is provided. The default port is `3939`. 

Two HTTP endpoints exist:

 - `/publish` which broadcasts the HTTP body to all subscribers.
 - `/subscribe` a blocking call, which receives a continuous stream of messages until the connection is terminated.

To run the server, a makefile is provided:
 - `make run`
 - To spawn subscribers: `curl -N localhost:3939/subscribe`
 - To broadcast messages: `curl localhost:3939/publish --data '1234'`

## Tests

Basic tests exists, covering most of the internals. Some more interesting test could be written modeling random connects/disconnects.

```
09:46 $ make test
go test github.com/ben-mays/coder-pubsub/pubsub
ok      github.com/ben-mays/coder-pubsub/pubsub 0.684s
go test -race github.com/ben-mays/coder-pubsub/pubsub
ok      github.com/ben-mays/coder-pubsub/pubsub 20.110s
go test -cover github.com/ben-mays/coder-pubsub/pubsub
ok      github.com/ben-mays/coder-pubsub/pubsub (cached)       coverage: 68.6% of statements
```
