package container

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/loheagn/cloud/docker"
)

type RunOption struct {
	Image   string
	Cmd     []string
	WorkDir string
	Mounts  map[string]string
}

func Run(ctx context.Context, opt *RunOption, output io.Writer) (exitCode int, err error) {
	cli, err := docker.GetDefaultClient()
	if err != nil {
		return 1, err
	}

	// 配置基本参数
	config := &container.Config{
		Image:      opt.Image,
		Cmd:        opt.Cmd,
		WorkingDir: opt.WorkDir,
	}
	// 挂载目录
	mounts := make([]mount.Mount, 0, len(opt.Mounts))
	for source, target := range opt.Mounts {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: source,
			Target: target,
		})
	}
	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return 1, err
	}

	// 保证最后将容器移除
	defer func() {
		_ = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	}()

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return 1, err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return 1, err
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return 1, err
	}

	_, err = stdcopy.StdCopy(output, output, out)
	if err != nil {
		return 1, err
	}

	status, err := cli.ContainerInspect(ctx, resp.ID)
	return status.State.ExitCode, err
}