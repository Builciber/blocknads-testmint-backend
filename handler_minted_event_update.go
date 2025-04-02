package main

import (
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func pollMintedEvent(cfg *apiConfig, tokenIdChan chan<- uint64, interval time.Duration) {
	tickChan := time.NewTicker(interval).C
	w, err := web3.NewWeb3(cfg.rpcUrl)
	w.Eth.SetChainId(cfg.chainID)
	if err != nil {
		log.Fatal(err.Error())
	}
	startBlock := cfg.wlMintStartBlock
	startBlockasHexStr := "0x" + strconv.FormatUint(startBlock, 16)
	toBlock := cfg.wlMintStartBlock + 9
	toBlockHexStr := "0x" + strconv.FormatUint(toBlock, 16)
	eventSignature := "Minted(address,uint256)"
	topicOne := crypto.Keccak256([]byte(eventSignature))
	log.Println("Started Minted event polling")
	for range tickChan {
		filter := types.Fliter{
			Address:   common.HexToAddress(cfg.contractAddress),
			FromBlock: startBlockasHexStr,
			ToBlock:   toBlockHexStr,
			Topics:    []string{"0x" + common.Bytes2Hex(topicOne)},
		}
		events, err := w.Eth.GetLogs(&filter)
		if err != nil {
			log.Fatalf("Error polling Minted events: %s", err.Error())
			continue
		}
		if len(events) == 0 {
			startBlock = toBlock + 1
			toBlock += 10
			startBlockasHexStr = "0x" + strconv.FormatUint(startBlock, 16)
			toBlockHexStr = "0x" + strconv.FormatUint(toBlock, 16)
			continue
		}
		latestEvent := events[len(events)-1]
		params, err := w.Utils.DecodeParameters([]string{"uint256"}, common.FromHex(latestEvent.Data))
		if err != nil {
			log.Fatalf("Error polling Minted events: %s", err.Error())
			continue
		}
		latestTokenId, ok := params[0].(*big.Int)
		if !ok {
			log.Fatalf("Error polling Minted events: Failed to parse latest `latestTokenId` as *big.Int")
			continue
		}
		tokenIdChan <- latestTokenId.Uint64()
		startBlock = toBlock + 1
		toBlock += 10
		startBlockasHexStr = "0x" + strconv.FormatUint(startBlock, 16)
		toBlockHexStr = "0x" + strconv.FormatUint(toBlock, 16)
	}
}

func (cfg *apiConfig) handlerMintedEventUpdate(tokenIdChan <-chan uint64) http.HandlerFunc {
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
			case tokenId := <-tokenIdChan:
				fmt.Fprintf(w, "event: mint\ndata: %d\n\n", tokenId+1)
				flusher.Flush()
			case <-r.Context().Done():
				return // Client disconnected
			}
		}
	}
	return http.HandlerFunc(fn)
}
