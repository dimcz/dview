package docker

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"terminal/internal/config"
	"terminal/pkg/logger"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Container struct {
	ID   string
	Name string
}

type Docker struct {
	containers []Container
	current    int

	cli *client.Client
	log *logger.Logger
	cfg *config.Config

	wg     *sync.WaitGroup
	ctx    context.Context
	cancel func()
}

func (d *Docker) Load(out io.Writer) error {
	d.ctx, d.cancel = context.WithCancel(context.Background())

	d.wg.Add(1)

	go func() {
		defer d.wg.Done()

		fd, err := d.cli.ContainerLogs(d.ctx, d.containers[d.current].ID, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Tail:       strconv.Itoa(d.cfg.Tail),
			Follow:     true,
		})
		if err != nil {
			d.log.Error("failed to load logs:", err)
		}

		defer func() {
			d.log.LogOnErr(fd.Close())
		}()

		if _, err := stdcopy.StdCopy(out, out, fd); err != nil {
			d.log.Error("failed StdCopy with err:", err)
			return
		}
	}()

	return nil
}

func (d *Docker) SetNextContainer() {
	c := d.current + 1
	if c >= len(d.containers) {
		c = 0
	}
	d.current = c
}

func (d *Docker) SetPrevContainer() {
	c := d.current - 1
	if c < 0 {
		c = len(d.containers) - 1
	}
	d.current = c
}

func (d *Docker) Name() string {
	return fmt.Sprintf("(%d/%d) %s (ID:%s)",
		d.current+1,
		len(d.containers),
		strings.Replace(d.containers[d.current].Name, "/", "", 1),
		d.containers[d.current].ID[:12])
}

func (d *Docker) Close() {
	d.Stop()

	d.log.LogOnErr(d.cli.Close())
}

func (d *Docker) Stop() {
	d.cancel()
	d.wg.Wait()
}

func getContainers(cli *client.Client) (containers []Container, err error) {
	list, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	for _, c := range list {
		containers = append(containers, Container{c.ID, strings.Join(c.Names, ", ")})
	}

	return containers, nil
}

func Client(log *logger.Logger, cfg *config.Config) (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	containers, err := getContainers(cli)
	if err != nil {
		return nil, err
	}

	return &Docker{
		log:        log,
		cfg:        cfg,
		cli:        cli,
		containers: containers,
		wg:         new(sync.WaitGroup),
	}, nil
}
