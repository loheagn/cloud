package docker

import (
	"bufio"
	"io"

	"github.com/docker/docker/client"
)

type InitOption struct {
	Host string
}

func GetClient(opt *InitOption) (cli *client.Client, err error) {
	if len(opt.Host) > 0 {
		cli, err = client.NewClientWithOpts(client.WithHost(opt.Host))
		if err == nil {
			return cli, err
		}
	}
	return client.NewClientWithOpts(client.FromEnv)
}

func GetDefaultClient() (cli *client.Client, err error) {
	return client.NewClientWithOpts(client.FromEnv)
}

func ReadOutput(rd io.Reader, output io.Writer) (err error) {
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		_, _ = output.Write(scanner.Bytes())
		_, _ = output.Write([]byte{'\n'})
	}
	return
}
