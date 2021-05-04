package server

type Process interface {
	Init() error
}

type ProcessRegisterFunc func(server *Server, id string, options map[string]interface{}) (Process, error)
