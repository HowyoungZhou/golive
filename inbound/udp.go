package inbound

import (
	"github.com/howyoungzhou/golive/server"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"net"
)

type UDPInboundOptions struct {
	Network string `json:"network"`
	Address string `json:"address"`
}

type UDPInbound struct {
	options *UDPInboundOptions
	logger  *log.Entry
	reader  *AsyncReader
}

func NewUDPInbound(options *UDPInboundOptions) (*UDPInbound, error) {
	res := &UDPInbound{
		options: options,
		logger:  log.New().WithFields(log.Fields{"module": "UDPInbound"}),
		reader:  NewAsyncReader(),
	}
	return res, nil
}

func RegisterUDPInbound(server *server.Server, id string, options map[string]interface{}) (server.Inbound, error) {
	opt := &UDPInboundOptions{}
	if err := mapstructure.Decode(options, opt); err != nil {
		return nil, err
	}
	return NewUDPInbound(opt)
}

func (s *UDPInbound) Init() error {
	conn, err := net.ListenPacket(s.options.Network, s.options.Address)
	if err != nil {
		s.logger.WithError(err).WithFields(log.Fields{"network": s.options.Network, "addr": s.options.Address}).Fatal("Failed to listen")
		return err
	}

	s.logger.WithFields(log.Fields{"network": s.options.Network, "addr": conn.LocalAddr()}).Info("The server is listening")
	go func() {
		defer conn.Close()
		for {
			n, _, err := conn.ReadFrom(s.reader.Fetch())
			s.reader.Return(n, err)
		}
	}()
	return nil
}

func (s *UDPInbound) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}
