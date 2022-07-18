package main

import (
	"fmt"
	"os"
	"terminal/internal/viewer"
	"terminal/pkg/docker"

	"terminal/internal/config"
	"terminal/pkg/logger"
)

const VERSION = "1.0.0"

func main() {
	cfg := config.Init()

	log := logger.Init(cfg.LogFile)
	defer log.Close()

	log.Info("Start DLOG version ", VERSION)

	if cfg.Version {
		fmt.Println("Dlog version", VERSION)
		os.Exit(0)
	}

	client, err := docker.Client(log, cfg)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	v, err := viewer.Init(log, cfg, client)
	if err != nil {
		log.Fatal(err)
	}

	defer v.Shutdown()

	if err := v.Start(); err != nil {
		log.Fatal(err)
	}
}
