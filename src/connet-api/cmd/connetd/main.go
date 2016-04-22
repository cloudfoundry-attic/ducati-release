package main

import (
	"connet-api/config"
	"connet-api/handlers"
	"connet-api/store"
	"encoding/json"
	"flag"
	"fmt"
	"lib/db"
	"lib/marshal"
	"log"
	"os"
	"time"

	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
	"github.com/tedsuo/rata"
)

const configFileFlag = "configFile"

func main() {
	var configFilePath string
	flag.StringVar(&configFilePath, configFileFlag, "", "")
	flag.Parse()

	conf, err := config.ParseConfigFile(configFilePath)
	if err != nil {
		log.Fatalf("parsing config: %s", err)
	}

	logger := lager.NewLogger("connetd")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

	retriableConnector := db.RetriableConnector{
		Connector:     db.GetConnectionPool,
		Sleeper:       db.SleeperFunc(time.Sleep),
		RetryInterval: 3 * time.Second,
		MaxRetries:    10,
	}

	databaseURL, err := conf.Database.PostgresURL()
	if err != nil {
		log.Fatalf("db config: %s", err)
	}

	dbConnectionPool, err := retriableConnector.GetConnectionPool(databaseURL)
	if err != nil {
		log.Fatalf("db connect: %s", err)
	}

	dataStore, err := store.New(dbConnectionPool)
	if err != nil {
		log.Fatalf("failed to construct datastore: %s", err)
	}

	rataHandlers := rata.Handlers{}

	rataHandlers["add_route"] = &handlers.AddRoute{
		Logger:      logger,
		Store:       dataStore,
		Unmarshaler: marshal.UnmarshalFunc(json.Unmarshal),
	}
	rataHandlers["list_routes"] = &handlers.ListRoutes{
		Logger:    logger,
		Store:     dataStore,
		Marshaler: marshal.MarshalFunc(json.Marshal),
	}

	routes := rata.Routes{
		{Name: "add_route", Method: "POST", Path: "/routes"},
		{Name: "list_routes", Method: "GET", Path: "/routes"},
	}

	rataRouter, err := rata.NewRouter(routes, rataHandlers)
	if err != nil {
		log.Fatalf("unable to create rata Router: %s", err) // not tested
	}

	httpServer := http_server.New(
		fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort),
		rataRouter,
	)

	members := grouper.Members{
		{"http_server", httpServer},
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	err = <-monitor.Wait()
	if err != nil {
		log.Fatalf("daemon terminated: %s", err)
	}
}
