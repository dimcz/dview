package main

import (
	"fmt"
	"os"

	"terminal/internal/config"
	"terminal/internal/terminal"
	"terminal/pkg/logger"
)

const VERSION = "1.0.0"

func main() {
	cfg := config.Init()

	log := logger.Init(cfg.LogFile)
	defer log.Close()

	log.Info("info", "Start DLOG version ", VERSION)

	if cfg.Version {
		fmt.Println("Dlog version", VERSION)
		os.Exit(0)
	}

	term, err := terminal.Init()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	defer term.Shutdown()
}
