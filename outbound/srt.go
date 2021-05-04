package outbound

import (
	"github.com/haivision/srtgo"
	"github.com/howyoungzhou/golive/server"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"sync"
)

type SRTOutboundOptions struct {
	Host       string
	Port       uint16
	Options    map[string]string
	BufferSize int
	Timeout    int
}

type SRTOutbound struct {
	options     *SRTOutboundOptions
	channels    map[string]chan []byte
	channelsMux sync.Mutex
	logger      *log.Entry
}

func NewSRTOutbound(options *SRTOutboundOptions) (*SRTOutbound, error) {
	return &SRTOutbound{
		options,
		make(map[string]chan []byte),
		sync.Mutex{},
		log.New().WithFields(log.Fields{"module": "SRTOutbound"}),
	}, nil
}

func RegisterSRTOutbound(server *server.Server, id string, options map[string]interface{}) (server.Outbound, error) {
	opt := &SRTOutboundOptions{}
	if err := mapstructure.Decode(options, opt); err != nil {
		return nil, err
	}
	return NewSRTOutbound(opt)
}

func (s *SRTOutbound) Init() error {
	sck := srtgo.NewSrtSocket(s.options.Host, s.options.Port, s.options.Options)

	err := sck.Listen(1)
	if err != nil {
		return err
	}
	s.logger.WithFields(log.Fields{"host": s.options.Host, "port": s.options.Port}).Info("The server is listening")

	go func() {
		for {
			cltSck, addr, err := sck.Accept()
			if err != nil {
				continue
			}
			s.logger.WithFields(log.Fields{"host": addr.IP, "port": addr.Port}).Info("Incoming connection")
			var channel = make(chan []byte, s.options.BufferSize)
			s.channelsMux.Lock()
			s.channels[addr.String()] = channel
			s.channelsMux.Unlock()
			go func() {
				for {
					data := <-channel
					_, err := cltSck.Write(data, s.options.Timeout)
					if err != nil {
						cltSck.Close()
						s.channelsMux.Lock()
						delete(s.channels, addr.String())
						s.channelsMux.Unlock()
						s.logger.WithFields(log.Fields{"host": addr.IP, "port": addr.Port}).Info("Connection closed")
						return
					}
				}
			}()
		}
	}()
	return nil
}

func (s *SRTOutbound) Write(data []byte) (int, error) {
	s.channelsMux.Lock()
	for addr, c := range s.channels {
		select {
		case c <- data:
		default:
			s.logger.WithFields(log.Fields{"addr": addr}).Warn("Connection blocked")
		}
	}
	s.channelsMux.Unlock()
	return len(data), nil
}
