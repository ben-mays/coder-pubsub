=== Pubsub

PubSub is a transport-agnostic in-memory pubsub library. A bi-directional `Client` can be registered with the pubsub broker, either before starting or during execution. Incoming messages on the `Client.In` channel will be broadcast to all clients. The broker will broadcast to all registered clients on the `Client.Out` channel.