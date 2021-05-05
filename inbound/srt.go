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

// SRTInbound implements SRT protocol for input
type SRTInbound struct {
	options *SRTInboundOptions
	logger  *log.Entry
	reader  *AsyncReader
}

// NewSrtpInbound creates a new instance of SRTInbound
func NewSrtpInbound(options *SRTInboundOptions) (*SRTInbound, error) {
	return &SRTInbound{
		options,
		log.New().WithFields(log.Fields{"module": "SRTInbound"}),
		NewAsyncReader(),
	}, nil
}

// RegisterSRTInbound registers a new instance to the server
func RegisterSRTInbound(server *server.Server, id string, options map[string]interface{}) (server.Inbound, error) {
	opt := &SRTInboundOptions{}
	if err := mapstructure.Decode(options, opt); err != nil {
		return nil, err
	}
	return NewSrtpInbound(opt)
}

// Init starts the SRT server
func (s *SRTInbound) Init() error {
	sck := srtgo.NewSrtSocket(s.options.Host, s.options.Port, s.options.Options)
	err := sck.Listen(2)
	if err != nil {
		panic(err)
	}
	s.logger.WithFields(log.Fields{"host": s.options.Host, "port": s.options.Port}).Info("The server is listening")
	go func() {
		// loop to accept new connections
		for {
			remoteSck, addr, err := sck.Accept()
			if err != nil {
				continue
			}
			s.logger.WithFields(log.Fields{"host": addr.IP, "port": addr.Port}).Info("Incoming connection")
			// loop to receive packets once the connection is established
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

// Read blocks to wait for a packet and puts it in the buffer
func (s *SRTInbound) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}
