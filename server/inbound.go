package server

type Inbound interface {
	Init() error
	Read(p []byte) (n int, err error)
}

type InboundRegisterFunc func(server *Server, id string, options map[string]interface{}) (Inbound, error)
