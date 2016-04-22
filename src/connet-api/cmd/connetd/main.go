package main

import (
	"connet-api/config"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
	"github.com/tedsuo/rata"
)

type MyHandler struct {
}

func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

const configFileFlag = "configFile"

func main() {
	var configFilePath string
	flag.StringVar(&configFilePath, configFileFlag, "", "")
	flag.Parse()

	conf, err := config.ParseConfigFile(configFilePath)
	if err != nil {
		log.Fatalf("parsing config: %s", err)
	}

	rataHandlers := rata.Handlers{}
	rataHandlers["hello"] = &MyHandler{}

	routes := rata.Routes{
		{Name: "hello", Method: "GET", Path: "/"},
	}

	rataRouter, err := rata.NewRouter(routes, rataHandlers)
	if err != nil {
		log.Fatalf("unable to create rata Router: %s", err) // not tested
	}

	httpServer := http_server.New(fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort), rataRouter)

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
