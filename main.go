package main

import (
	"context"
	"flag"

	"github.com/samkreter/go-core/log"
	"github.com/samkreter/givedirectly/apiserver"
)

var (
	logLvl        string

	serverConfig = &apiserver.Config{}
)

func main() {
	flag.StringVar(&logLvl, "log-level", "info", "the log level for the application")

	flag.StringVar(&serverConfig.ServerAddr, "addr", "0.0.0.0:8080", "the address to expose the API server")
	flag.BoolVar(&serverConfig.EnableReqLogging, "enable-req-logging", true, "Enable logging for all incoming requests")
	flag.BoolVar(&serverConfig.EnableReqCorrelation, "enable-req-corr", true, "Enable correlation for all incoming requests")
	flag.Parse()

	ctx := context.Background()
	logger := log.G(ctx)

	if err := log.SetLogLevel(logLvl); err != nil {
		logger.Errorf("failed to set log level to : '%s'", logLvl)
	}

	server, err := apiserver.NewServer(serverConfig)
	if err != nil {
		logger.Fatal(err)
	}

	if err = server.Run(); err != nil {
		logger.Fatal(err)
	}
}
