package main

import (
	"github.com/subsan/http2privat_json/pkg/config"
	"github.com/subsan/http2privat_json/pkg/connector"
	"github.com/subsan/http2privat_json/pkg/webserver"
	"log"
	"strconv"
)

func main() {
	err := config.Load()
	if err != nil {
		log.Fatalf(" [EE] Cannot load config: %v", err.Error())
	}

	go webserver.Start(config.Config.Webserver.Port)
	go connector.Listener(config.Config.Terminal.Ip + ":" + strconv.Itoa(config.Config.Terminal.Port))

	select {}
}
