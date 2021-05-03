package server

import "errors"

type Server struct {
	registeredInbound  map[string]InboundRegisterFunc
	registeredOutbound map[string]OutboundRegisterFunc
	inbounds           map[string]Inbound
	outbounds          map[string]Outbound
	pipes              map[Inbound][]Outbound
}

type InboundRegisterFunc func(server *Server, id string, options map[string]interface{}) (Inbound, error)

type OutboundRegisterFunc func(server *Server, id string, options map[string]interface{}) (Outbound, error)

func New() *Server {
	return &Server{
		make(map[string]InboundRegisterFunc),
		make(map[string]OutboundRegisterFunc),
		make(map[string]Inbound),
		make(map[string]Outbound),
		make(map[Inbound][]Outbound),
	}
}

func (s *Server) RegisterInbound(name string, regFunc InboundRegisterFunc) {
	s.registeredInbound[name] = regFunc
}

func (s *Server) AddInboundObj(id string, o Inbound) {
	s.inbounds[id] = o
}

func (s *Server) AddInbound(id, typ string, options map[string]interface{}) error {
	f, ok := s.registeredInbound[typ]
	if !ok {
		return errors.New("unknown inbound: " + typ)
	}
	o, err := f(s, id, options)
	if err != nil {
		return err
	}
	s.AddInboundObj(id, o)
	return nil
}

func (s *Server) RegisterOutbound(name string, regFunc OutboundRegisterFunc) {
	s.registeredOutbound[name] = regFunc
}

func (s *Server) AddOutboundObj(id string, o Outbound) {
	s.outbounds[id] = o
}

func (s *Server) AddOutbound(id, typ string, options map[string]interface{}) error {
	f, ok := s.registeredOutbound[typ]
	if !ok {
		return errors.New("unknown outbound: " + typ)
	}
	o, err := f(s, id, options)
	if err != nil {
		return err
	}
	s.AddOutboundObj(id, o)
	return nil
}

func (s *Server) AddPipe(in string, outs []string) error {
	i, ok := s.inbounds[in]
	if !ok {
		return errors.New("unknown inbound: " + in)
	}
	var outbounds []Outbound
	for _, out := range outs {
		o, ok := s.outbounds[out]
		if !ok {
			return errors.New("unknown outbounds: " + out)
		}
		outbounds = append(outbounds, o)
	}
	s.pipes[i] = outbounds
	return nil
}

func (s *Server) Run() error {
	for _, i := range s.inbounds {
		err := i.Init()
		if err != nil {
			return err
		}
	}

	for _, o := range s.outbounds {
		err := o.Init()
		if err != nil {
			return err
		}
	}

	for in, outs := range s.pipes {
		go func() {
			for {
				buffer := make([]byte, 10240)
				n, err := in.Read(buffer)
				if n == 0 {
					continue
				}
				if err != nil {
					// TODO: handle error
					panic(err)
				}
				for _, out := range outs {
					_, err := out.Write(buffer[:n])
					if err != nil {
						panic(err)
					}
				}
			}
		}()
	}
	return nil
}
