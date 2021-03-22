package main

import (
	"context"
	"flag"
	"github.com/samkreter/givedirectly/datastore"
	"github.com/samkreter/givedirectly/types"
	"time"

	"github.com/samkreter/givedirectly/apiserver"
	"github.com/samkreter/go-core/log"
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

	// Ensure there's enough time for the postgres db to initialize. In prod, i'd use either retries or if it's deployed
	// to Kubernetes, let the pod restarts handle it.
	time.Sleep(time.Second * 3)

	sqlStore, err := datastore.NewSQLStore(pgUser, pgDBName, pgPassword, pgHost, pgPort)
	if err != nil {
		logger.Fatal(err)
	}

	// Create all the required tables in the DB
	if err := sqlStore.EnsureDB(); err != nil {
		logger.Fatal(err)
	}

	testBooks := []*types.Book{
		{Available: true, Title: "testbook"},
		{Available: true, Title: "testbook2"},
		{Available: false, Title: "testbook3"},
	}

	// Seed the DB with some books. Use 3 known title books plus many randomly generated books
	if err := sqlStore.SeedDB(numToSeed, testBooks...); err != nil {
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
