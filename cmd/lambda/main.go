package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dkoshkin/invoices-validator/pkg/controller"
	log "github.com/sirupsen/logrus"
	"os"
)

func HandleRequest() error {
	return controller.Run()
}

const logLevelEnv = "LOG_LEVEL"

func main() {

	// setup logging
	logLevel := os.Getenv(logLevelEnv)

	if logLevel != "" {
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			log.Errorf("error parsing log level, will fallback to default level: %v", err)
		} else {
			log.SetLevel(level)
		}
	}

	lambda.Start(HandleRequest)
}
