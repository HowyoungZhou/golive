package outbound

import (
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/howyoungzhou/golive/server"
	"github.com/mitchellh/mapstructure"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type WebRTCOutboundOptions struct {
	WebRTC webrtc.Configuration
	Tracks []struct {
		CodecCapability webrtc.RTPCodecCapability
		ID              string
		StreamID        string
	}
	SDPServer struct {
		CORS          cors.Config
		ListenAddress string
		RootPath      string
	}
}

// WebRTCOutbound implements WebRTC protocol for output
type WebRTCOutbound struct {
	options *WebRTCOutboundOptions
	tracks  []*webrtc.TrackLocalStaticRTP
	logger  *log.Entry
}

// NewWebRTCOutbound creates a new instance of WebRTCOutbound
func NewWebRTCOutbound(options *WebRTCOutboundOptions) (*WebRTCOutbound, error) {
	res := &WebRTCOutbound{
		options: options,
		logger:  log.New().WithFields(log.Fields{"module": "WebRTCOutbound"}),
	}
	for _, t := range options.Tracks {
		track, err := webrtc.NewTrackLocalStaticRTP(t.CodecCapability, t.ID, t.StreamID)
		if err != nil {
			return nil, err
		}
		res.tracks = append(res.tracks, track)
	}
	return res, nil
}

// RegisterWebRTC registers a new instance to the server, create a new sub-outbound for each track
func RegisterWebRTC(server *server.Server, id string, options map[string]interface{}) (server.Outbound, error) {
	opt := &WebRTCOutboundOptions{}
	if err := mapstructure.Decode(options, opt); err != nil {
		return nil, err
	}
	res, err := NewWebRTCOutbound(opt)
	if err != nil {
		return nil, err
	}
	for _, t := range res.tracks {
		server.AddWriter(id+":"+t.ID(), &WebRtcTrackOutbound{t})
	}
	return res, nil
}

func (o *WebRTCOutbound) handleSDPRequest(c *gin.Context) {
	offer := webrtc.SessionDescription{}
	err := c.Bind(&offer)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		o.logger.WithField("addr", c.Request.RemoteAddr).Info("malformed SDP")
		return
	}
	peerConnectionConfig := o.options.WebRTC
	// Create a new PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		o.logger.WithField("addr", c.Request.RemoteAddr).Error("failed to create peer connection")
		return
	}

	for _, track := range o.tracks {
		rtpSender, err := peerConnection.AddTrack(track)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			o.logger.WithField("addr", c.Request.RemoteAddr).Info("failed to add track")
			return
		}

		// Read incoming RTCP packets
		// Before these packets are returned they are processed by interceptors. For things
		// like NACK this needs to be called.
		go func() {
			rtcpBuf := make([]byte, 1500)
			for {
				if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
					return
				}
			}
		}()
	}
	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		o.logger.WithField("addr", c.Request.RemoteAddr).Info("failed to set remote SDP")
		return
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		o.logger.WithField("addr", c.Request.RemoteAddr).Info("failed to create answer SDP")
		return
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		o.logger.WithField("addr", c.Request.RemoteAddr).Info("failed to set local SDP")
		return
	}

	<-gatherComplete

	// Get the LocalDescription and take it to base64 so we can paste in browser
	c.JSON(http.StatusOK, peerConnection.LocalDescription())
	o.logger.WithField("addr", c.Request.RemoteAddr).Info("connection established")
}

func (o *WebRTCOutbound) serveHTTP() {
	r := gin.Default()
	r.Use(cors.New(o.options.SDPServer.CORS))
	r.POST(o.options.SDPServer.RootPath, o.handleSDPRequest)
	o.logger.WithField("addr", o.options.SDPServer.ListenAddress).Info("SDP server is listening")
	err := r.Run(o.options.SDPServer.ListenAddress)
	o.logger.WithField("addr", o.options.SDPServer.ListenAddress).WithError(err).Error("SDP server ended with error")
}

// Init runs the HTTP SDP server
func (o *WebRTCOutbound) Init() error {
	go o.serveHTTP()
	return nil
}

func (o *WebRTCOutbound) Write(p []byte) (int, error) {
	return 0, errors.New("can not write directly to a WebRTC outbound, change \"out\" to \"[outbound id]:[track id]\" instead")
}

type WebRtcTrackOutbound struct {
	track *webrtc.TrackLocalStaticRTP
}

func (o *WebRtcTrackOutbound) Init() error {
	return nil
}

// Write writes packet to the track
func (o *WebRtcTrackOutbound) Write(p []byte) (int, error) {
	return o.track.Write(p)
}
