package connector

import (
	json2 "encoding/json"
	"log"
	"sync"
	"time"

	"github.com/subsan/http2privat_json/pkg/config"
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

	if err := ensureConnected(); err != nil {
		inTransaction.mu.Lock()
		inTransaction.v = false
		inTransaction.mu.Unlock()

		return JsonEntity{
			Error:            true,
			ErrorDescription: err.Error(),
		}
	}
	stopKeepAlive()

	// Drop anything left in the buffer from previous transactions
	// (late responses, stale ServiceMessage, etc.) — otherwise the next
	// request would receive a response that does not belong to it.
	drainBuffer()

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

	expectedMethod := json.Method
	deadline := time.After(config.Config.Timeout.Transaction)

	for {
		select {
		case answer := <-buffer:
			// Match the response by Method. Anything that does not match
			// (intermediate ServiceMessage, late frames) is logged and
			// skipped while we keep waiting for the real response.
			if answer.Method != expectedMethod {
				log.Printf(" [WW] [connector] [syncSender] skipping unmatched message (expected method=%q, got method=%q): %+v\n",
					expectedMethod, answer.Method, answer)
				continue
			}
			inTransaction.mu.Lock()
			inTransaction.v = false
			inTransaction.mu.Unlock()
			startOrResetKeepAlive()
			return answer
		case <-deadline:
			log.Printf(" [WW] [connector] [syncSender] Timeout waiting response message (expected method=%q)", expectedMethod)
			interrupt()
			// Safety net: drop anything that may have arrived right now,
			// so the next SyncSender starts with a clean channel.
			drainBuffer()
			inTransaction.mu.Lock()
			inTransaction.v = false
			inTransaction.mu.Unlock()
			startOrResetKeepAlive()
			return JsonEntity{
				Error:            true,
				ErrorDescription: "Transaction timeout",
			}
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
	log.Printf(" [  ] [connector] [sender] Initialize sender: %+v\n", json)
	if err := ensureConnected(); err != nil {
		return err
	}
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

	startOrResetKeepAlive()

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
