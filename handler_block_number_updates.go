package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/chenzhijie/go-web3"
)

func pollBlockNumber(cfg *apiConfig, broadcastChan chan<- uint64, interval time.Duration) {
	tickChan := time.NewTicker(interval).C
	w, err := web3.NewWeb3(cfg.rpcUrl)
	w.Eth.SetChainId(cfg.chainID)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("Started polling block numbers")
	for range tickChan {
		blockNumber, err := w.Eth.GetBlockNumber()
		if err != nil {
			continue
		}
		broadcastChan <- blockNumber
	}
}

func (cfg *apiConfig) handlerBlockUpdates(broadcastChan <-chan uint64) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		for {
			select {
			case blockNumber := <-broadcastChan:
				fmt.Fprintf(w, "event: blocknumber\ndata: %d\n\n", blockNumber)
				flusher.Flush()
			case <-r.Context().Done():
				return // Client disconnected
			}
		}
	}
	return http.HandlerFunc(fn)
}
