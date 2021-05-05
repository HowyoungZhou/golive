package inbound

// AsyncReader implements io.Reader interface with asynchronous read performed with channel
type AsyncReader struct {
	bufChannel chan []byte
	nChannel   chan int
	errChannel chan error
}

// NewAsyncReader creates a new instance of AsyncReader
func NewAsyncReader() *AsyncReader {
	return &AsyncReader{
		make(chan []byte),
		make(chan int),
		make(chan error),
	}
}

// Read blocks until read request completes
func (r *AsyncReader) Read(p []byte) (n int, err error) {
	r.bufChannel <- p
	return <-r.nChannel, <-r.errChannel
}

// Fetch blocks to wait a new read request
func (r *AsyncReader) Fetch() []byte {
	return <-r.bufChannel
}

// Return responds to the latest request
func (r *AsyncReader) Return(n int, err error) {
	r.nChannel <- n
	r.errChannel <- err
}
