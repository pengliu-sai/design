package design

import (
	"os"
	"log"
)

type Options struct {
	HTTPAddress string `flag:"http-address"`

	Logger logger
}

func NewOptions() *Options {
	return &Options{
		HTTPAddress: "0.0.0.0:8080",
		Logger: log.New(os.Stderr, "[design] ", log.Ldate|log.Ltime|log.Lmicroseconds),
	}
}
