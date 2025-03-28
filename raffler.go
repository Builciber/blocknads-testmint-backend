package main

import (
	"context"
	"log"
	"math"
	"math/big"
	"math/rand/v2"
	"net/http"
	"slices"

	"github.com/Builciber/blocknads-testmint-backend/internal/database"
	"github.com/chenzhijie/go-web3"
)

type walletAddress string

func weightedRandomSelect(addressWeightGrouping [][]walletAddress, cumWeights []int, totalWeight int) walletAddress {
	randVal := rand.IntN(totalWeight)
	var groupIdx int
	for i, cumWeight := range cumWeights {
		if randVal < cumWeight {
			groupIdx = i
			break
		}
	}
	groupLength := len(addressWeightGrouping[groupIdx])
	selectedWalletIdx := rand.IntN(groupLength)
	selectedWallet := addressWeightGrouping[groupIdx][selectedWalletIdx]
	temp := addressWeightGrouping[groupIdx][groupLength-1]
	addressWeightGrouping[groupIdx][groupLength-1] = selectedWallet
	addressWeightGrouping[groupIdx][selectedWalletIdx] = temp
	addressWeightGrouping[groupIdx] = slices.Delete(addressWeightGrouping[groupIdx], groupLength-1, groupLength)
	return selectedWallet
}

func (cfg *apiConfig) raffler(w http.ResponseWriter, r *http.Request) {
	ticketBuyers, err := cfg.DB.GetAllTicketBuyers(context.Background())
	if err != nil {
		log.Fatal(err.Error())
	}
	weights, err := cfg.DB.GetUniqueWeights(context.Background())
	if err != nil {
		log.Fatal(err.Error())
	}
	addressWeightGrouping := make([][]walletAddress, len(weights))
	weight := weights[0]
	weightIdx := 0
	for _, ticketBuyer := range ticketBuyers {
		if ticketBuyer.NumTickets == weight {
			addressWeightGrouping[weightIdx] = append(addressWeightGrouping[weightIdx], walletAddress(ticketBuyer.WalletAddress))
			continue
		}
		weight = ticketBuyer.NumTickets
		weightIdx++
		addressWeightGrouping[weightIdx] = append(addressWeightGrouping[weightIdx], walletAddress(ticketBuyer.WalletAddress))
	}
	totalWeight := 0
	cumWeights := make([]int, len(weights))
	for i, w := range weights {
		totalWeight += int(w)
		cumWeights[i] = totalWeight
	}
	wb, err := web3.NewWeb3(cfg.rpcUrl)
	if err != nil {
		log.Fatal(err.Error())
	}
	wb.Eth.SetChainId(cfg.chainID)
	contract, err := wb.Eth.NewContract(abi, cfg.contractAddress)
	if err != nil {
		log.Fatal(err.Error())
	}
	hex, err := contract.Call("MAX_NFTS")
	if err != nil {
		log.Fatal(err.Error())
	}
	totalSupply, ok := hex.(*big.Int)
	if !ok {
		log.Fatal("Failed to parse `totalSupply` as *big.Int")
	}
	totalSupplyAsInt := totalSupply.Int64()
	slotsSize := math.Ceil(float64(totalSupplyAsInt) / 256)
	slots := make([]*big.Int, int(slotsSize))
	for i := 0; i < int(slotsSize); i++ {
		index := big.NewInt(int64(i))
		val, err := contract.Call("slots", index)
		if err != nil {
			log.Fatal(err.Error())
		}
		slotVal, ok := val.(*big.Int)
		if !ok {
			log.Fatal("Failed to parse `slotVal` as *big.Int")
		}
		slots[i] = slotVal
	}
	unusedNonces := []uint{}
	for i, slot := range slots {
		bitLength := slot.BitLen()
		log.Println("Bit length of slot is: ", bitLength)
		for j := 0; i < bitLength; i++ {
			nonce := uint(i*256 + j)
			if slot.Bit(j) == 1 && nonce < uint(totalSupplyAsInt) {
				unusedNonces = append(unusedNonces, nonce)
			}
		}
	}
	numRaffleWinners := len(unusedNonces)
	raffleWinners := make([]database.CreateRaffleWinnersForTxParams, numRaffleWinners)
	for i := 0; i < numRaffleWinners; i++ {
		winnerAddress := weightedRandomSelect(addressWeightGrouping, cumWeights, totalWeight)
		raffleWinners[i] = database.CreateRaffleWinnersForTxParams{
			WalletAddress: string(winnerAddress),
			Nonce:         int16(unusedNonces[i]),
		}
	}

	err = cfg.updateTicketBuyersNonceTx(context.Background(), raffleWinners)
	if err != nil {
		log.Fatal(err.Error())
	}
	w.WriteHeader(http.StatusOK)
}
