package config

import "github.com/spf13/pflag"

type Config struct {
	Version bool
	LogFile string
	Tail    int
	Mouse   bool
}

func Init() *Config {
	var config Config

	pflag.BoolVarP(&(config.Version),
		"version", "v", false, "Print version information.")
	pflag.BoolVarP(&(config.Mouse),
		"mouse", "m", false, "Mouse support.")
	pflag.StringVarP(&(config.LogFile),
		"log", "l", "", "Send log messages to file.")
	pflag.IntVarP(&(config.Tail),
		"tail", "t", 1_000, "Number of lines to show from the end of the logs.")
	pflag.Parse()

	return &config
}
