package image

import (
	"context"
	"path/filepath"
	"testing"
)

func TestBuild(t *testing.T) {
	type args struct {
		ctx context.Context
		opt *BuildOption
	}
	dockerfileAbsPath, err := filepath.Abs(filepath.Join("../example/ubuntu-test", "./Dockerfile"))
	if err != nil {
		panic(err.Error())
	}
	ctxAbsPath, err := filepath.Abs("../example/ubuntu-test")
	if err != nil {
		panic(err.Error())
	}

	tests := []struct {
		name        string
		args        args
		checkOutput bool
		wantErr     bool
	}{
		{
			name: "local-build",
			args: args{
				ctx: context.Background(),
				opt: &BuildOption{
					DockerFilePath: "./Dockerfile",
					CtxPath:        "../example/ubuntu-test",
					Tags:           []string{"test/ubuntu:20.04"},
				},
			},
		},
		{
			name: "local-build-dockerfile-abspath",
			args: args{
				ctx: context.Background(),
				opt: &BuildOption{
					DockerFilePath: dockerfileAbsPath,
					CtxPath:        "../example/ubuntu-test",
					Tags:           []string{"test/ubuntu:20.04"},
				},
			},
			wantErr: true,
		},
		{
			name: "local-build-context-abspath",
			args: args{
				ctx: context.Background(),
				opt: &BuildOption{
					DockerFilePath: "./Dockerfile",
					CtxPath:        ctxAbsPath,
					Tags:           []string{"test/ubuntu:20.04"},
				},
			},
		},
		{
			name: "local-build-path-test",
			args: args{
				ctx: context.Background(),
				opt: &BuildOption{
					DockerFilePath: "./ubuntu-test/Dockerfile",
					CtxPath:        "../example/",
					Tags:           []string{"test/ubuntu:20.04"},
				},
			},
		},

		{
			name: "remote-build",
			args: args{
				ctx: context.Background(),
				opt: &BuildOption{
					HostURL:        "tcp://10.251.0.45:2375",
					DockerFilePath: "./Dockerfile",
					CtxPath:        "../example/ubuntu-test",
					Tags:           []string{"test/ubuntu:20.04"},
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Build(tt.args.ctx, tt.args.opt); (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
