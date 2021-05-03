package server

type Outbound interface {
	Init() error
	Write(p []byte) (int, error)
}
