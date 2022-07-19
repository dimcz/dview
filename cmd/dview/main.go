package main

import (
	"fmt"

	"github.com/dimcz/viewer/internal/config"
	"github.com/dimcz/viewer/internal/viewer"
	"github.com/dimcz/viewer/pkg/docker"
	"github.com/dimcz/viewer/pkg/logger"
)

const VERSION = "0.0.6"

func main() {
	cfg := config.Init()

	log := logger.Init(cfg.LogFile)
	defer log.Close()

	log.Info("Start DVIEW version ", VERSION)

	if cfg.Version {
		fmt.Println("DView version", VERSION)

		return
	}

	client, err := docker.Client(log, cfg)
	if err != nil {
		fmt.Println(err)

		return
	}

	defer client.Close()

	v, err := viewer.Init(log, cfg, client)
	if err != nil {
		fmt.Println(err)

		return
	}

	defer v.Shutdown()

	if err := v.Start(); err != nil {
		fmt.Println(err)
	}
}
