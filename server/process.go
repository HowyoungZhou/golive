package server

type Process interface {
	Init() error
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
}

type ProcessRegisterFunc func(server *Server, id string, options map[string]interface{}) (Process, error)
