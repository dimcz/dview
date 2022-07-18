package docker

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

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

	wg     *sync.WaitGroup
	ctx    context.Context
	cancel func()
}

func (d *Docker) Load(out io.Writer) {
	d.ctx, d.cancel = context.WithCancel(context.Background())

	info, err := d.cli.ContainerInspect(d.ctx, d.containers[d.current].ID)
	if err != nil {
		return
	}

	opts := types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: false,
		Tail:       strconv.Itoa(d.cfg.Tail),
		Follow:     false,
	}

	if err := d.download(info.Config.Tty, out, opts); err != nil {
		d.log.Error(err)
	}

	d.wg.Add(1)

	go func() {
		defer d.wg.Done()

		opts.Tail = "0"
		opts.Follow = true

		if err := d.download(info.Config.Tty, out, opts); err != nil {
			d.log.Error("failed to download with: ", err)
		}
	}()
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

func (d *Docker) download(tty bool, out io.Writer, opts types.ContainerLogsOptions) error {
	fd, err := d.cli.ContainerLogs(d.ctx, d.containers[d.current].ID, opts)
	if err != nil {
		d.log.Error("failed to load logs:", err)
	}

	defer func() {
		d.log.LogOnErr(fd.Close())
	}()

	if tty {
		_, err = io.Copy(out, fd)
	} else {
		_, err = stdcopy.StdCopy(out, out, fd)
	}

	return err
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
