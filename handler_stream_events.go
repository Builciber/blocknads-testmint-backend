package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type event struct {
	data               uint64
	isBlockNumberEvent bool
}

type client struct {
	eventChan chan event
}

func broadcastEvent(clients map[*client]struct{}, mutex *sync.Mutex, event event) {
	defer mutex.Unlock()
	mutex.Lock()
	for client := range clients {
		select {
		case client.eventChan <- event:
		default:
			log.Println("Dropping update for slow client")
		}
	}
}

var currBlockNumber uint64

func (cfg *apiConfig) handlerStreamEvents(clients map[*client]struct{}, clientMu *sync.Mutex) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			log.Println("Streaming unsupported")
			return
		}
		client := client{eventChan: make(chan event, 2)}
		clientMu.Lock()
		clients[&client] = struct{}{}
		clientMu.Unlock()
		totalMinted, err := cfg.DB.GetTotalMinted(r.Context())
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}
		fmt.Fprintf(w, "event: mint\ndata: %d\n\n", totalMinted)
		flusher.Flush()
		if currBlockNumber > 0 {
			fmt.Fprintf(w, "event: blocknumber\ndata: %d\n\n", currBlockNumber)
			flusher.Flush()
		}
		for {
			select {
			case event := <-client.eventChan:
				if event.isBlockNumberEvent {
					fmt.Fprintf(w, "event: blocknumber\ndata: %d\n\n", event.data)
					flusher.Flush()
				} else {
					fmt.Fprintf(w, "event: mint\ndata: %d\n\n", event.data)
					flusher.Flush()
				}
			case <-r.Context().Done(): // Client disconnected
				clientMu.Lock()
				delete(clients, &client)
				clientMu.Unlock()
				close(client.eventChan)
				return
			}
		}
	}
	return http.HandlerFunc(fn)
}

func pollBlockNumber(cfg *apiConfig, clients map[*client]struct{}, clientMu *sync.Mutex, interval time.Duration, raffleStartChan chan<- uint64, hasRaffled *bool) {
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
		currBlockNumber = blockNumber
		broadcastEvent(clients, clientMu, event{
			data:               blockNumber,
			isBlockNumberEvent: true,
		})
		if *hasRaffled {
			continue
		}
		raffleStartChan <- blockNumber
	}
}

func pollMintedEvent(cfg *apiConfig, clients map[*client]struct{}, clientMu *sync.Mutex, interval time.Duration) {
	tickChan := time.NewTicker(interval).C
	w, err := web3.NewWeb3(cfg.rpcUrl)
	w.Eth.SetChainId(cfg.chainID)
	if err != nil {
		log.Fatal(err.Error())
	}
	startBlock, err := w.Eth.GetBlockNumber()
	if err != nil {
		log.Fatal(err.Error())
	}
	startBlockasHexStr := "0x" + strconv.FormatUint(startBlock, 16)
	eventSignature := "Minted(address,uint256)"
	topicOne := crypto.Keccak256([]byte(eventSignature))
	log.Println("Started polling Minted event")
	for range tickChan {
		toBlock, err := w.Eth.GetBlockNumber()
		if err != nil {
			log.Fatal(err.Error())
		}
		toBlockHexStr := "0x" + strconv.FormatUint(toBlock, 16)
		filter := types.Fliter{
			Address:   common.HexToAddress(cfg.contractAddress),
			FromBlock: startBlockasHexStr,
			ToBlock:   toBlockHexStr,
			Topics:    []string{"0x" + common.Bytes2Hex(topicOne)},
		}
		events, err := w.Eth.GetLogs(&filter)
		if err != nil {
			log.Printf("Error polling Minted events: %s", err.Error())
			continue
		}
		if len(events) == 0 {
			startBlock = toBlock + 1
			startBlockasHexStr = "0x" + strconv.FormatUint(startBlock, 16)
			continue
		}
		latestEvent := events[len(events)-1]
		params, err := w.Utils.DecodeParameters([]string{"uint256"}, common.FromHex(latestEvent.Data))
		if err != nil {
			log.Fatalf("Error polling Minted events: %s", err.Error())
		}
		latestTokenId, ok := params[0].(*big.Int)
		if !ok {
			log.Printf("Error polling Minted events: Failed to parse latest `latestTokenId` as *big.Int")
			continue
		}
		broadcastEvent(clients, clientMu, event{
			data:               latestTokenId.Uint64() + 1,
			isBlockNumberEvent: false,
		})
		startBlock = toBlock + 1
		startBlockasHexStr = "0x" + strconv.FormatUint(startBlock, 16)
		err = cfg.DB.UpdateTotalNftsMinted(context.Background(), int16(latestTokenId.Int64())+1)
		if err != nil {
			log.Printf("Error polling Minted events: %s", err.Error())
		}
	}
}
