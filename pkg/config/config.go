package config

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Struct struct {
	Webserver struct {
		Port int `default:"3333"`
	}
	Terminal struct {
		Ip   string `default:"127.0.0.1"`
		Port int    `default:"2000"`
	}
	Timeout struct {
		Connection  time.Duration `default:"15s"`
		Write       time.Duration `default:"15s"`
		Reconnect   time.Duration `default:"5s"`
		Transaction time.Duration `default:"10m"`
		KeepAlive   time.Duration `default:"60s"`
	}
}

var Config Struct

func readEnv() (err error) {
	err = envconfig.Process("", &Config)
	if err != nil {
		return err
	}

	return nil
}

func Load() (err error) {
	err = godotenv.Load()
	if err != nil {
		log.Printf(" [WW] [config] %v\n", err)
	}

	err = readEnv()
	if err != nil {
		return err
	}

	return nil
}
