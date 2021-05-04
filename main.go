package main

import (
	"encoding/json"
	"flag"
	"github.com/howyoungzhou/golive/inbound"
	"github.com/howyoungzhou/golive/outbound"
	"github.com/howyoungzhou/golive/process"
	"github.com/howyoungzhou/golive/server"
	"io/ioutil"
)

type Options struct {
	Inbounds []struct {
		Id      string `json:"id"`
		Type    string `json:"type"`
		Options map[string]interface{}
	} `json:"inbounds"`
	Outbounds []struct {
		Id      string `json:"id"`
		Type    string `json:"type"`
		Options map[string]interface{}
	} `json:"outbounds"`
	Processes []struct {
		Id      string `json:"id"`
		Type    string `json:"type"`
		Options map[string]interface{}
	} `json:"processes"`
	Pipes []struct {
		In   string   `json:"in"`
		Outs []string `json:"outs"`
	} `json:"pipes"`
}

func main() {
	configPath := flag.String("config", "config.json", "path to the config file")
	flag.Parse()
	configData, err := ioutil.ReadFile(*configPath)
	if err != nil {
		panic(err)
	}
	options := Options{}
	err = json.Unmarshal(configData, &options)
	if err != nil {
		panic(err)
	}

	s := server.New()
	s.RegisterInbound("udp", inbound.RegisterUDPInbound)
	s.RegisterInbound("tcp", inbound.RegisterTCPInbound)
	s.RegisterInbound("srt", inbound.RegisterSRTInbound)
	s.RegisterOutbound("webrtc", outbound.RegisterWebRTC)
	s.RegisterOutbound("srt", outbound.RegisterSRTOutbound)
	s.RegisterProcess("exec", process.RegisterExecProcess)

	for _, i := range options.Inbounds {
		err := s.AddInbound(i.Id, i.Type, i.Options)
		if err != nil {
			panic(err)
		}
	}

	for _, o := range options.Outbounds {
		err := s.AddOutbound(o.Id, o.Type, o.Options)
		if err != nil {
			panic(err)
		}
	}

	for _, p := range options.Processes {
		err := s.AddProcess(p.Id, p.Type, p.Options)
		if err != nil {
			panic(err)
		}
	}

	for _, pipe := range options.Pipes {
		err := s.AddPipe(pipe.In, pipe.Outs)
		if err != nil {
			panic(err)
		}
	}
	s.Run()
	for {
	}
}
