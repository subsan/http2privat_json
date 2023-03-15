package connector

import (
	"bytes"
	json2 "encoding/json"
	"github.com/subsan/http2privat_json/pkg/config"
	"io"
	"log"
	"net"
	"time"
)

var connection net.Conn
var isConnected = false
var connectionAddress string
var buffer = make(chan JsonEntity, 1)

func connect() {
	var err error
	dialer := net.Dialer{Timeout: config.Config.Timeout.Connection}
	connection, err = dialer.Dial("tcp", connectionAddress)
	if err != nil {
		isConnected = false
		log.Printf(" [WW] [connector] Connect error: %+v", err)
	} else {
		isConnected = true
		log.Printf(" [  ] [connector] Connected to: %+v", connectionAddress)
	}
}

func disconnect() {
	isConnected = false

	if connection != nil {
		err := connection.Close()
		if err != nil {
			log.Printf(" [  ] [connector] error when disconnect")

			return
		}
	}
}

func reconnect() {
	disconnect()

	for !isConnected {
		time.Sleep(config.Config.Timeout.Reconnect)
		log.Printf(" [  ] [connector] try reconnect")
		disconnect()
		connect()
	}
}

func Listener(address string) {
	connectionAddress = address
	connect()

	if !isConnected {
		reconnect()
	}

	var err error

	answer := make([]byte, 1024)
	buf := bytes.Buffer{}
	for {
		i := 0
		i, err = connection.Read(answer)
		if err != nil && err != io.EOF {
			log.Printf(" [WW] [connector] error reading message: %+v\n", err)
			reconnect()

			break
		}
		if i != 0 {
			answer = answer[:i]
			buf.Write(answer)
			if buf.Bytes()[len(buf.Bytes())-1] == 0x00 {
				a := buf.Bytes()[:len(buf.Bytes())-1]

				var json JsonEntity
				err := json2.Unmarshal(a, &json)
				if err != nil {
					log.Printf(" [WW] [connector] Error unmarshaling answer JSON: %+v\n", err)
					reconnect()

					break
				}
				log.Printf(" [  ] [connector] Get message: %+v\n", json)
				buffer <- json
				buf.Reset()
			}
		}
	}
}
