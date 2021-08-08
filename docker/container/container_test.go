package container

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loheagn/loclo/docker/image"
)

func Test_Run(t *testing.T) {
	const tag string = "test/ubuntu:20.04"
	const dockerDURL = "tcp://10.251.0.45:2375"
	mountTestPath, err := filepath.Abs("../example/mount-test")
	if err != nil {
		panic(err.Error())
	}
	// 先创建一下这个测试用的基础镜像
	remoteBuildOpt := &image.BuildOption{
		HostURL:        dockerDURL,
		DockerFilePath: "./Dockerfile",
		CtxPath:        "../example/ubuntu-test",
		Tags:           []string{tag},
	}
	_, _ = image.Build(context.Background(), remoteBuildOpt)

	envBuildOpt := &image.BuildOption{
		DockerFilePath: "./Dockerfile",
		CtxPath:        "../example/ubuntu-test",
		Tags:           []string{tag},
	}
	_, _ = image.Build(context.Background(), envBuildOpt)

	type args struct {
		image  string
		config *RunOption
	}
	tests := []struct {
		name       string
		args       args
		exitNormal bool
		wantErr    bool

		checkOutput bool
		output      string
	}{
		{
			args: args{
				image: tag,
				config: &RunOption{
					Image: tag,
					Cmd:   []string{"./error.sh"},
				},
			},
			exitNormal: false,
		},

		{
			args: args{
				image: tag,
				config: &RunOption{
					Image: tag,
					Cmd:   []string{"./success.sh"},
				},
			},
			exitNormal: true,
		},

		{
			name: "workdir-test",
			args: args{
				image: tag,
				config: &RunOption{
					Image:   tag,
					WorkDir: "/etc/apt",
					Cmd:     []string{"pwd"},
				},
			},
			exitNormal:  true,
			checkOutput: true,
			output:      "/etc/apt",
		},

		{
			name: "mount-test",
			args: args{
				image: tag,
				config: &RunOption{
					Image:   tag,
					WorkDir: "/etc/apt",
					Cmd:     []string{"ls", "/tmp"},
					Mounts: map[string]string{
						mountTestPath: "/tmp",
					},
				},
			},
			exitNormal:  true,
			checkOutput: true,
			output:      "data",
		},

		{
			name: "host-url-test",
			args: args{
				image: tag,
				config: &RunOption{
					HostURL: dockerDURL,
					Image:   tag,
					Cmd:     []string{"uname"},
				},
			},
			exitNormal:  true,
			checkOutput: true,
			output:      "Linux",
		},

		{
			name: "env-test",
			args: args{
				image: tag,
				config: &RunOption{
					Image: tag,
					Envs: map[string]string{
						"CONTAINER_TEST_ENV": "test-env",
					},
					Cmd: []string{"bash", "-c", "echo $CONTAINER_TEST_ENV"},
				},
			},
			exitNormal:  true,
			checkOutput: true,
			output:      "test-env",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, exitCode, err := Run(context.TODO(), tt.args.config)
			if (err != nil) == tt.wantErr && (exitCode == 0) == tt.exitNormal {
				if !tt.checkOutput || (output == tt.output || strings.TrimSpace(output) == tt.output) {
					return
				}
				t.Errorf("Run() writer = %s, wantOutput %s", output, tt.output)
			}
			t.Log(output)
			t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			t.Errorf("Run() exitCode = %v, exitNormal %v", exitCode, tt.exitNormal)
		})
	}
}
