package main

import (
	"testing"
	"os"
	"github.com/BurntSushi/toml"
	"framework/design"
	"github.com/mreiferson/go-options"
)

func TestConfigFlagParsing(t *testing.T) {
	flagSet := designFlagSet()
	flagSet.Parse([]string{})

	var cfg config
	f, err := os.Open("./config/design.cfg")
	if err != nil {
		t.Fatalf("%s", err)
	}

	toml.DecodeReader(f, &cfg)

	opts := design.NewOptions()
	options.Resolve(opts, flagSet, cfg)
	design.New(opts)
}
