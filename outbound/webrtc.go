package outbound

import (
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/howyoungzhou/golive/server"
	"github.com/mitchellh/mapstructure"
	"github.com/pion/webrtc/v3"
	"net/http"
	"time"
)

type WebRTCOutboundOptions struct {
	webrtc.Configuration
	Tracks []struct {
		CodecCapability webrtc.RTPCodecCapability
		ID              string
		StreamID        string
	}
}

type WebRTCOutbound struct {
	options *WebRTCOutboundOptions
	tracks  []*webrtc.TrackLocalStaticRTP
}

func NewWebRTCOutbound(options *WebRTCOutboundOptions) (*WebRTCOutbound, error) {
	res := &WebRTCOutbound{
		options: options,
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

func (o *WebRTCOutbound) serveHTTP() {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))
	r.POST("/", func(c *gin.Context) {
		offer := webrtc.SessionDescription{}
		err := c.Bind(&offer)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		// TODO: load from config
		peerConnectionConfig := webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		}
		// Create a new PeerConnection
		peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfig)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		for _, track := range o.tracks {
			rtpSender, err := peerConnection.AddTrack(track)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
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
			return
		}

		// Create answer
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Create channel that is blocked until ICE Gathering is complete
		gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

		// Sets the LocalDescription, and starts our UDP listeners
		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		<-gatherComplete

		// Get the LocalDescription and take it to base64 so we can paste in browser
		c.JSON(http.StatusOK, peerConnection.LocalDescription())
	})
	// TODO: load config from option
	r.Run("0.0.0.0:8080")
}

func (o *WebRTCOutbound) Init() error {
	go o.serveHTTP()
	return nil
}

func (o *WebRTCOutbound) Write(p []byte) (int, error) {
	return 0, errors.New("can not write directly to a WebRTC outbound, change \"out\" to [outbound id]:[track id] instead")
}

type WebRtcTrackOutbound struct {
	track *webrtc.TrackLocalStaticRTP
}

func (o *WebRtcTrackOutbound) Init() error {
	return nil
}

func (o *WebRtcTrackOutbound) Write(p []byte) (int, error) {
	return o.track.Write(p)
}
