package inbound

import (
	"github.com/haivision/srtgo"
	"github.com/howyoungzhou/golive/server"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
)

type SRTInboundOptions struct {
	Host    string
	Port    uint16
	Timeout int
	Options map[string]string
}

type SRTInbound struct {
	options *SRTInboundOptions
	logger  *log.Entry
	reader  *AsyncReader
}

func NewSrtpInbound(options *SRTInboundOptions) (*SRTInbound, error) {
	return &SRTInbound{
		options,
		log.New().WithFields(log.Fields{"module": "SRTInbound"}),
		NewAsyncReader(),
	}, nil
}

func RegisterSRTInbound(server *server.Server, id string, options map[string]interface{}) (server.Inbound, error) {
	opt := &SRTInboundOptions{}
	if err := mapstructure.Decode(options, opt); err != nil {
		return nil, err
	}
	return NewSrtpInbound(opt)
}

func (s *SRTInbound) Init() error {
	sck := srtgo.NewSrtSocket(s.options.Host, s.options.Port, s.options.Options)
	err := sck.Listen(2)
	if err != nil {
		panic(err)
	}
	s.logger.WithFields(log.Fields{"host": s.options.Host, "port": s.options.Port}).Info("The server is listening")
	go func() {
		for {
			remoteSck, addr, err := sck.Accept()
			if err != nil {
				continue
			}
			s.logger.WithFields(log.Fields{"host": addr.IP, "port": addr.Port}).Info("Incoming connection")
			for {
				n, err := remoteSck.Read(s.reader.Fetch(), s.options.Timeout)
				s.reader.Return(n, err)
				if err != nil {
					remoteSck.Close()
					s.logger.WithFields(log.Fields{"host": addr.IP, "port": addr.Port}).Info("Connection closed")
					break
				}
			}
		}
	}()
	return nil
}

func (s *SRTInbound) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}
