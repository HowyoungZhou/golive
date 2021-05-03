package inbound

type AsyncReader struct {
	bufChannel chan []byte
	nChannel   chan int
	errChannel chan error
}

func NewAsyncReader() *AsyncReader {
	return &AsyncReader{
		make(chan []byte),
		make(chan int),
		make(chan error),
	}
}

func (r *AsyncReader) Read(p []byte) (n int, err error) {
	r.bufChannel <- p
	return <-r.nChannel, <-r.errChannel
}

func (r *AsyncReader) Fetch() []byte {
	return <-r.bufChannel
}

func (r *AsyncReader) Return(n int, err error) {
	r.nChannel <- n
	r.errChannel <- err
}
