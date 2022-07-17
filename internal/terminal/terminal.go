package terminal

import (
	"context"

	"terminal/pkg/docker"
)

type Terminal struct {
	ctx    context.Context
	cancel func()

	dock *docker.Docker
}

func Init() (*Terminal, error) {
	ctx, cancel := context.WithCancel(context.Background())

	d, err := docker.Client(ctx, file)
	if err != nil {
		return nil, err
	}

	return &Terminal{ctx, cancel}
}

func (t *Terminal) Shutdown() {
	t.cancel()
}
