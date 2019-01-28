package main

import (
	"flag"
	"github.com/dkoshkin/invoices-validator/pkg/controller"
	log "github.com/sirupsen/logrus"
)

func main() {
	// setup logging
	logLevel := flag.String("log-level", "info", "log level")
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Errorf("error parsing log level, will fallback to default level: %v", err)
	} else {
		log.SetLevel(level)
	}

	controller.Run()
}
