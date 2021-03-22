package main

import (
	"context"
	"flag"
	"github.com/samkreter/givedirectly/datastore"

	"github.com/samkreter/go-core/log"
	"github.com/samkreter/givedirectly/apiserver"
)

var (
	logLvl        string
	pgHost, pgUser, pgPassword, pgDBName string
	pgPort int

	numToSeed int

	serverConfig = &apiserver.Config{}
)

func main() {
	flag.StringVar(&logLvl, "log-level", "info", "the log level for the application")

	flag.StringVar(&serverConfig.ServerAddr, "addr", "0.0.0.0:8080", "the address to expose the API server")
	flag.BoolVar(&serverConfig.EnableReqLogging, "enable-req-logging", true, "Enable logging for all incoming requests")
	flag.BoolVar(&serverConfig.EnableReqCorrelation, "enable-req-corr", true, "Enable correlation for all incoming requests")

	// Postgres configuration
	flag.StringVar(&pgUser, "pg-user", "librarystore", "the postgres user")
	flag.StringVar(&pgPassword, "pg-password", "", "the postgres password")
	flag.StringVar(&pgDBName, "pg-dbname", "librarystore", "the postgres dbname")
	flag.StringVar(&pgHost, "pg-host", "0.0.0.0", "the postgres host")
	flag.IntVar(&pgPort, "pg-port", 5432, "the postgres port")
	flag.IntVar(&numToSeed, "seednum", 100, "the number of books to seed the db")

	flag.Parse()

	ctx := context.Background()
	logger := log.G(ctx)

	if err := log.SetLogLevel(logLvl); err != nil {
		logger.Errorf("failed to set log level to : '%s'", logLvl)
	}

	sqlStore, err := datastore.NewSQLStore(pgUser, pgDBName, pgPassword, pgHost, pgPort)
	if err != nil {
		logger.Fatal(err)
	}

	if err := sqlStore.EnsureDB(); err != nil {
		logger.Fatal(err)
	}

	if err := sqlStore.SeedDB(numToSeed); err != nil {
		logger.Fatal(err)
	}

	server, err := apiserver.NewServer(sqlStore, serverConfig)
	if err != nil {
		logger.Fatal(err)
	}

	if err = server.Run(); err != nil {
		logger.Fatal(err)
	}
}
