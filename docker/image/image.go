package image

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/loheagn/cloud/docker"
)

type BuildOption struct {
	HostURL        string
	DockerFilePath string
	CtxPath        string
	Tags           []string
}

func Build(ctx context.Context, opt *BuildOption, output io.Writer) error {
	cli, err := docker.GetClient(&docker.InitOption{
		Host: opt.HostURL,
	})
	if err != nil {
		return err
	}
	buildOpts := types.ImageBuildOptions{
		Dockerfile: opt.DockerFilePath,
		Tags:       opt.Tags,
	}
	buildCtx, _ := archive.TarWithOptions(opt.CtxPath, &archive.TarOptions{})

	resp, err := cli.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(output, resp.Body)
	return err
}
