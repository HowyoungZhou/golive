package inbound

import (
	"github.com/howyoungzhou/golive/server"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"net"
)

type TCPInboundOptions struct {
	Network string `json:"network"`
	Address string `json:"address"`
}

type TCPInbound struct {
	options *TCPInboundOptions
	logger  *log.Entry
	reader  *AsyncReader
}

func NewTCPInbound(options *TCPInboundOptions) (*TCPInbound, error) {
	res := &TCPInbound{
		options: options,
		logger:  log.New().WithFields(log.Fields{"module": "TCPInbound"}),
		reader:  NewAsyncReader(),
	}
	return res, nil
}

func RegisterTCPInbound(server *server.Server, id string, options map[string]interface{}) (server.Inbound, error) {
	opt := &TCPInboundOptions{}
	if err := mapstructure.Decode(options, opt); err != nil {
		return nil, err
	}
	return NewTCPInbound(opt)
}

func (s *TCPInbound) Init() error {
	ln, err := net.Listen(s.options.Network, s.options.Address)
	if err != nil {
		s.logger.WithError(err).WithFields(log.Fields{"network": s.options.Network, "addr": s.options.Address}).Fatal("Failed to listen")
		return err
	}

	s.logger.WithFields(log.Fields{"network": s.options.Network, "addr": ln.Addr()}).Info("The server is listening")
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				s.logger.WithField("addr", conn.RemoteAddr()).Warn("Failed to accept connection")
				continue
			}
			s.logger.WithField("addr", conn.RemoteAddr()).Info("Incoming connection")
			for {
				n, err := conn.Read(s.reader.Fetch())
				s.reader.Return(n, err)
				if err != nil {
					conn.Close()
					s.logger.WithFields(log.Fields{"addr": conn.RemoteAddr()}).Info("Connection closed")
					break
				}
			}
		}
	}()
	return nil
}

func (s *TCPInbound) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}
