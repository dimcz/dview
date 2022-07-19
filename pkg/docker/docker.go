package docker

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/dimcz/viewer/internal/config"
	"github.com/dimcz/viewer/pkg/logger"
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

	ctx    context.Context
	cancel func()
}

func (d *Docker) Load(ctx context.Context, out io.Writer, tail int) {
	d.ctx, d.cancel = context.WithCancel(ctx)

	info, err := d.cli.ContainerInspect(d.ctx, d.containers[d.current].ID)
	if err != nil {
		return
	}

	opts := types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: d.cfg.Timestamp,
		Follow:     true,
	}

	if tail > 0 {
		opts.Tail = strconv.Itoa(tail)
	}

	go d.download(info.Config.Tty, out, opts)
	time.Sleep(100 * time.Millisecond)
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
}

func (d *Docker) download(tty bool, out io.Writer, opts types.ContainerLogsOptions) {
	fd, err := d.cli.ContainerLogs(d.ctx, d.containers[d.current].ID, opts)
	if err != nil {
		d.log.Error("failed to load logs:", err)
	}

	defer func() {
		d.log.LogOnErr(fd.Close())
	}()

	if tty {
		_, _ = io.Copy(out, fd)
	} else {
		_, _ = stdcopy.StdCopy(out, out, fd)
	}
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
	}, nil
}
