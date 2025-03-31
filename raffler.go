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

func weightedRandomSelect(addressWeightGrouping [][]walletAddress, weights []int16, cumWeights []int, totalWeight *int) walletAddress {
	randVal := rand.IntN(*totalWeight)
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
	if len(addressWeightGrouping[groupIdx]) == 0 {
		addressWeightGrouping = slices.Delete(addressWeightGrouping, groupIdx, groupIdx+1)
		weights = slices.Delete(weights, groupIdx, groupIdx+1)
		cumWeights = slices.Delete(cumWeights, groupIdx, groupIdx+1)
		sumWeight := 0
		for i, w := range weights {
			sumWeight += int(w)
			cumWeights[i] = sumWeight
		}
		*totalWeight = sumWeight
	}
	return selectedWallet
}

func (cfg *apiConfig) raffler(w http.ResponseWriter, r *http.Request) {
	ticketBuyers, err := cfg.DB.GetAllTicketBuyers(context.Background())
	if err != nil {
		log.Fatal(err.Error())
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
	totalSupply, ok := hex.(uint16)
	if !ok {
		log.Fatal("Failed to parse `totalSupply` as uint16")
	}
	slotsSize := math.Ceil(float64(totalSupply) / 256)
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
		for j := 0; j < bitLength; j++ {
			nonce := uint(i*256 + j)
			if slot.Bit(j) == 1 && nonce < uint(totalSupply) {
				unusedNonces = append(unusedNonces, nonce)
			}
		}
	}
	// Handle the case where the total number of ticket buyers is less than or equal to the number of
	// NFTs available for the raffle mint
	if len(ticketBuyers) <= len(unusedNonces) {
		numRaffleWinners := len(ticketBuyers)
		raffleWinners := make([]database.CreateRaffleWinnersForTxParams, numRaffleWinners)
		for i := 0; i < numRaffleWinners; i++ {
			raffleWinners[i] = database.CreateRaffleWinnersForTxParams{
				WalletAddress: ticketBuyers[i].WalletAddress,
				Nonce:         int16(unusedNonces[i]),
			}
		}
		err = cfg.updateTicketBuyersNonceTx(context.Background(), raffleWinners)
		if err != nil {
			log.Fatal(err.Error())
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	// Handle the case where the total number of ticket buyers is greater than the number of
	// NFTs available for the raffle mint
	weights, err := cfg.DB.GetUniqueWeights(context.Background())
	if err != nil {
		log.Fatal(err.Error())
	}
	addressWeightGrouping := make([][]walletAddress, len(weights))
	for i := 0; i < len(addressWeightGrouping); i++ {
		addressWeightGrouping[i] = make([]walletAddress, 0)
	}
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
	numRaffleWinners := len(unusedNonces)
	raffleWinners := make([]database.CreateRaffleWinnersForTxParams, numRaffleWinners)
	for i := 0; i < numRaffleWinners; i++ {
		winnerAddress := weightedRandomSelect(addressWeightGrouping, weights, cumWeights, &totalWeight)
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
