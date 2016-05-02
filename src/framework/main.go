package main

import (
	"flag"
	"fmt"
	"framework/design"
	"framework/design/tools/version"
	"github.com/BurntSushi/toml"
	"github.com/mreiferson/go-options"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func designFlagSet() *flag.FlagSet {
	flagSet := flag.NewFlagSet("design", flag.ExitOnError)

	flagSet.String("config", "", "path to config file")
	flagSet.Bool("version", false, "print version string")
	flagSet.String("http-address", "0.0.0.0:8080", "<addr>:<port> to listen on for HTTP clients")

	return flagSet
}

type config map[string]interface{}

func main() {
	flagSet := designFlagSet()
	flagSet.Parse(os.Args[1:])

	if flagSet.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("design"))
		os.Exit(0)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	var cfg config
	configFile := flagSet.Lookup("config").Value.String()
	if configFile != "" {
		_, err := toml.DecodeFile(configFile, &cfg)
		if err != nil {
			log.Fatalf("ERROR: failed to load config file %s - %s", configFile, err.Error())
		}
	}

	opts := design.NewOptions()
	options.Resolve(opts, flagSet, cfg)
	design := design.New(opts)

	design.Main()

	<-signalChan
	design.Exit()
}
