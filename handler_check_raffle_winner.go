package main

import (
	"log"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
)

type checkWinnerResp struct {
	WonRaffle bool `json:"won_raffle"`
}

func (cfg *apiConfig) handlerCheckRaffleWinner(w http.ResponseWriter, r *http.Request) {
	walletAddress := chi.URLParam(r, "walletAddress")
	if ok, _ := regexp.MatchString(`^0x[0-9a-fA-F]{40}$`, walletAddress); !ok {
		http.Error(w, "Invalid wallet address", http.StatusBadRequest)
		log.Println("Error at `handlerCheckWinner`: Invalid wallet address")
		return
	}
	exists, err := cfg.DB.IsRaffleWinner(r.Context(), walletAddress)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("Error at `handlerCheckWinner`: %s", err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, checkWinnerResp{WonRaffle: exists})
}
