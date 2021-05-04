package server

type Outbound interface {
	Init() error
	Write(p []byte) (int, error)
}

type OutboundRegisterFunc func(server *Server, id string, options map[string]interface{}) (Outbound, error)
