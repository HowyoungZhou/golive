package process

import (
	"bufio"
	"github.com/howyoungzhou/golive/server"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

type ExecProcessOptions struct {
	Path string
	Args []string
}

type ExecProcess struct {
	options *ExecProcessOptions
}

func NewExecProcess(options *ExecProcessOptions) (*ExecProcess, error) {
	return &ExecProcess{options: options}, nil
}

func RegisterExecProcess(server *server.Server, id string, options map[string]interface{}) (server.Process, error) {
	opt := &ExecProcessOptions{}
	if err := mapstructure.Decode(options, opt); err != nil {
		return nil, err
	}
	return NewExecProcess(opt)
}

func (e *ExecProcess) Init() error {
	cmd := exec.Command(e.options.Path, e.options.Args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		in := bufio.NewScanner(stderr)

		for in.Scan() {
			log.Errorln(in.Text())
		}
		if err := in.Err(); err != nil {
			log.WithError(err).Error("failed to read stderr")
		}
	}()

	go func() {
		in := bufio.NewScanner(stdout)

		for in.Scan() {
			log.Infoln(in.Text())
		}
		if err := in.Err(); err != nil {
			log.WithError(err).Error("failed to read stdout")
		}
	}()
	return nil
}
