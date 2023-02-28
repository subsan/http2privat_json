package webserver

import (
	"fmt"
	"log"
	"net/http"
)

func Start(port int) {
	http.HandleFunc("/check", getCheck)
	http.HandleFunc("/command", command)

	log.Printf(" [  ] [webserver] Listen on port: %d", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("Cannot start webserver: %+v", err)

		return
	}
}
