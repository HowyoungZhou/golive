package server

type Inbound interface {
	Init() error
	Read(p []byte) (n int, err error)
}
