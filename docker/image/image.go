package image

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/loheagn/loclo/docker"
)

type BuildOption struct {
	HostURL        string
	DockerFilePath string
	CtxPath        string
	Tags           []string
}

func Build(ctx context.Context, opt *BuildOption) (string, error) {
	cli, err := docker.GetClient(&docker.InitOption{
		Host: opt.HostURL,
	})
	if err != nil {
		return "", err
	}
	buildOpts := types.ImageBuildOptions{
		Dockerfile: opt.DockerFilePath,
		Tags:       opt.Tags,
	}
	buildCtx, _ := archive.TarWithOptions(opt.CtxPath, &archive.TarOptions{})

	resp, err := cli.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	return handleOutput(resp.Body)
}

type PushOption struct {
	HostURL  string
	Tag      string
	Username string
	Password string
}

func Push(ctx context.Context, opt *PushOption) (string, error) {
	cli, err := docker.GetClient(&docker.InitOption{
		Host: opt.HostURL,
	})
	if err != nil {
		return "", err
	}

	authConfig := types.AuthConfig{
		Username: opt.Username,
		Password: opt.Password,
	}
	encodedJSON, _ := json.Marshal(authConfig)
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	resp, err := cli.ImagePush(ctx, opt.Tag, types.ImagePushOptions{
		All:           false,
		RegistryAuth:  authStr,
		PrivilegeFunc: nil,
	})
	if err != nil {
		return "", err
	}
	defer func(resp io.ReadCloser) {
		_ = resp.Close()
	}(resp)
	return handleOutput(resp)
}

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func handleOutput(reader io.Reader) (string, error) {
	var (
		lastLine string
		buff     bytes.Buffer
	)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lastLine = scanner.Text()
		buff.WriteString(lastLine)
		_ = buff.WriteByte('\n')
	}

	errLine := &ErrorLine{}
	_ = json.Unmarshal([]byte(lastLine), errLine)
	if errLine.Error != "" {
		return buff.String(), errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
		return buff.String(), err
	}
	return buff.String(), nil
}
