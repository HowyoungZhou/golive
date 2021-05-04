package process

import (
	"bufio"
	"github.com/howyoungzhou/golive/server"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
)

type ExecProcessOptions struct {
	Path string
	Args []string
}

type ExecProcess struct {
	options *ExecProcessOptions
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
}

func NewExecProcess(options *ExecProcessOptions) (*ExecProcess, error) {
	return &ExecProcess{
		options: options,
		cmd:     exec.Command(options.Path, options.Args...),
	}, nil
}

func RegisterExecProcess(server *server.Server, id string, options map[string]interface{}) (server.Process, error) {
	opt := &ExecProcessOptions{}
	if err := mapstructure.Decode(options, opt); err != nil {
		return nil, err
	}
	return NewExecProcess(opt)
}

func (e *ExecProcess) Init() error {
	var err error
	e.stdin, err = e.cmd.StdinPipe()
	if err != nil {
		return err
	}

	e.stdout, err = e.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := e.cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = e.cmd.Start()
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

	return nil
}

func (e *ExecProcess) Read(p []byte) (n int, err error) {
	return e.stdout.Read(p)
}

func (e *ExecProcess) Write(p []byte) (n int, err error) {
	return e.stdin.Write(p)
}
