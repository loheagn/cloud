package kube

import (
	"context"
	"testing"
)

func TestPodDeploy(t *testing.T) {
	type args struct {
		ctx context.Context
		opt *PodDeployOpt
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "simple-test",
			args: args{
				ctx: context.Background(),
				opt: &PodDeployOpt{
					KubeConfPath: "",
					Labels: map[string]string{
						"simple": "test",
					},
					ReplicaNum: 0,
					Stateful:   false,
					Namespace:  "",
					Duration:   0,
					spec: PodSpec{
						Name:     "simple-test",
						ImageTag: "nginx:1.17",
						Envs:     nil,
						Ports: []Port{
							{
								Name: "main",
								Port: 80,
							},
						},
						WorkDir: "",
						Cmd:     Cmd{},
						Labels: map[string]string{
							"simple": "test",
							"inner":  "pod",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PodDeploy(tt.args.ctx, tt.args.opt); (err != nil) != tt.wantErr {
				t.Errorf("PodDeploy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
