package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/dimcz/viewer/internal/config"
	"github.com/dimcz/viewer/internal/viewer"
	"github.com/dimcz/viewer/pkg/docker"
	"github.com/dimcz/viewer/pkg/logger"
)

const VERSION = "1.0.0"

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	cfg := config.Init()

	log := logger.Init(cfg.LogFile)
	defer log.Close()

	log.Info("Start DLOG version ", VERSION)

	if cfg.Version {
		fmt.Println("Dlog version", VERSION)

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

		return
	}
}
