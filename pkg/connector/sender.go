package connector

import (
	json2 "encoding/json"
	"github.com/subsan/http2privat_json/pkg/config"
	"log"
	"sync"
	"time"
)

type inTransactionStructure struct {
	mu sync.Mutex
	v  bool
}

var inTransaction inTransactionStructure

func SyncSender(json JsonEntity) JsonEntity {
	log.Printf(" [  ] [connector] [sender] Initialize sync sender: %+v\n", json)
	inTransaction.mu.Lock()
	if inTransaction.v {
		inTransaction.mu.Unlock()
		return JsonEntity{
			Error:            true,
			ErrorDescription: "Another transaction active",
		}
	}
	inTransaction.v = true
	inTransaction.mu.Unlock()

	err := sender(json)
	if err != nil {
		inTransaction.mu.Lock()
		inTransaction.v = false
		inTransaction.mu.Unlock()

		return JsonEntity{
			Error:            true,
			ErrorDescription: err.Error(),
		}
	}

	select {
	case answer := <-buffer:
		inTransaction.mu.Lock()
		inTransaction.v = false
		inTransaction.mu.Unlock()

		return answer
	case <-time.After(config.Config.Timeout.Transaction):
		log.Printf(" [WW] [connector] [syncSender] Timeout waiting response message")
		interrupt()
		inTransaction.mu.Lock()
		inTransaction.v = false
		inTransaction.mu.Unlock()

		return JsonEntity{
			Error:            true,
			ErrorDescription: "Transaction timeout",
		}
	}
}

func interrupt() {
	log.Printf(" [  ] [connector] [sender] Sending interrupt message")

	err := sender(JsonEntity{
		Method: "ServiceMessage",
		Params: map[string]string{
			"msgType": "interrupt",
		},
	})
	if err != nil {
		log.Printf(" [WW] [connector] [sender] Error sending interrupt message: %+v\n", err)
	}
}

func sender(json JsonEntity) error {
	data, err := json2.Marshal(json)
	if err != nil {
		log.Printf(" [WW] [connector] [sender] Error marshaling JSON: %+v\n", err)

		return err
	}
	data = append(data, 0x00)
	err = writeEth(data)
	if err != nil {
		log.Printf(" [WW] [connector] [sender] Error conn.Write: %+v\n", err)

		return err
	}

	return nil
}

func writeEth(buffer []byte) error {
	var n int
	var err error

	if err = connection.SetWriteDeadline(time.Now().Add(config.Config.Timeout.Write)); err != nil {
		log.Printf(" [WW] [connector] [writter] Error set deadline: %+v\n", err)

		return err
	}

	n, err = connection.Write(buffer)
	if err != nil {
		log.Printf(" [WW] [connector] [writter] Error TCP write: %+v\n", err)

		return err
	}

	log.Printf(" [  ] [connector] [sender] wrote %d bytes.", n)
	return err
}
