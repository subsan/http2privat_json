package connector

import (
	"bytes"
	json2 "encoding/json"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/subsan/http2privat_json/pkg/config"
)

var connection net.Conn
var isConnected = false
var connectionAddress string
var buffer = make(chan JsonEntity, 1)
var connMu sync.Mutex
var keepAliveMu sync.Mutex
var keepAliveTimer *time.Timer

func stopKeepAlive() {
	keepAliveMu.Lock()
	defer keepAliveMu.Unlock()
	if keepAliveTimer != nil {
		if !keepAliveTimer.Stop() {
			select {
			case <-keepAliveTimer.C:
			default:
			}
		}
		keepAliveTimer = nil
	}
}

func startOrResetKeepAlive() {
	keepAliveMu.Lock()
	defer keepAliveMu.Unlock()
	if keepAliveTimer == nil {
		keepAliveTimer = time.NewTimer(config.Config.Timeout.KeepAlive)
		go func(t *time.Timer) {
			<-t.C
			disconnect()
		}(keepAliveTimer)
		return
	}
	if !keepAliveTimer.Stop() {
		select {
		case <-keepAliveTimer.C:
		default:
		}
	}
	keepAliveTimer.Reset(config.Config.Timeout.KeepAlive)
}

func connect() {
	connMu.Lock()
	defer connMu.Unlock()
	var err error
	dialer := net.Dialer{Timeout: config.Config.Timeout.Connection}
	connection, err = dialer.Dial("tcp", connectionAddress)
	if err != nil {
		isConnected = false
		log.Printf(" [WW] [connector] Connect error: %+v", err)
	} else {
		isConnected = true
		log.Printf(" [  ] [connector] Connected to: %+v", connectionAddress)
		go readLoop()
	}
}

func disconnect() {
	stopKeepAlive()
	connMu.Lock()
	defer connMu.Unlock()
	isConnected = false
	if connection != nil {
		err := connection.Close()
		if err != nil {
			log.Printf(" [  ] [connector] error when disconnect")
			return
		}
		connection = nil
	}
}

func ensureConnected() error {
	if isConnected {
		return nil
	}
	connect()
	if !isConnected {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func Listener(address string) {
	connectionAddress = address
}

func readLoop() {
	var err error

	answer := make([]byte, 1024)
	buf := bytes.Buffer{}
	for {
		i := 0
		connMu.Lock()
		c := connection
		connMu.Unlock()
		if c == nil {
			return
		}
		i, err = c.Read(answer)
		if err != nil && err != io.EOF {
			log.Printf(" [WW] [connector] error reading message: %+v\n", err)
			connMu.Lock()
			isConnected = false
			if connection != nil {
				_ = connection.Close()
				connection = nil
			}
			connMu.Unlock()
			return
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
					disconnect()
					return
				}
				log.Printf(" [  ] [connector] Get message: %+v\n", json)
				buffer <- json
				buf.Reset()
			}
		}
	}
}
