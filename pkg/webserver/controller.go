package webserver

import (
	json2 "encoding/json"
	"github.com/golang/gddo/httputil/header"
	"github.com/subsan/http2privat_json/pkg/connector"
	"log"
	"net/http"
)

func getCheck(w http.ResponseWriter, r *http.Request) {
	response := connector.CheckConnection()
	log.Printf(" [  ] [webserver] [check]: %+v\n", response)

	json, _ := json2.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(json)
	if err != nil {
		log.Printf(" [EE] [webserver] [check] Error write into response: %+v", err.Error())
	}
}

func command(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			log.Printf(" [EE] [webserver] [command]: %+v", msg)
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	var request connector.JsonEntity

	err := json2.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf(" [EE] [webserver] [command] Error parse json: %+v", err.Error())

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf(" [  ] [webserver] [command] Executing: %+v\n", request)

	response := connector.SyncSender(request)

	log.Printf(" [  ] [webserver] [command] response: %+v\n", response)

	json, _ := json2.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(json)
	if err != nil {
		log.Printf(" [EE] [webserver] [command] Error write into response: %+v", err.Error())
	}
}
