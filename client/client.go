package client

// Client is a simple container for bi-directional communication. It could be further extended to include state.
type Client struct {
	In  chan []byte
	Out chan []byte
}
