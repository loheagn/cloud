package container

import (
	"bytes"
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/loheagn/loclo/docker"
)

type RunOption struct {
	HostURL   string
	Image     string
	Cmd       []string
	Envs      map[string]string
	WorkDir   string
	Mounts    map[string]string
	Resources *container.Resources
}

func Run(ctx context.Context, opt *RunOption) (output string, exitCode int, err error) {
	cli, err := docker.GetClient(&docker.InitOption{
		Host: opt.HostURL,
	})
	if err != nil {
		return "", 1, err
	}

	// 配置基本参数
	// TODO: 参数核验
	envs := make([]string, 0, len(opt.Envs))
	for k, v := range opt.Envs {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	config := &container.Config{
		Image:      opt.Image,
		Cmd:        opt.Cmd,
		WorkingDir: opt.WorkDir,
		Env:        envs,
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
		Mounts:    mounts,
		Resources: *opt.Resources,
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", 1, err
	}

	// 保证最后将容器移除
	defer func() {
		_ = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	}()

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", 1, err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", 1, err
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return "", 1, err
	}

	var writer = &bytes.Buffer{}
	_, err = stdcopy.StdCopy(writer, writer, out)
	if err != nil {
		return "", 1, err
	}

	status, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", 1, err
	}
	return writer.String(), status.State.ExitCode, err
}
