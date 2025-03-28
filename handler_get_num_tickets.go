package main

import (
	"log"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type getNumTicketsResp struct {
	NumTickets int `json:"num_tickets"`
}

func (cfg *apiConfig) handlerGetNumTickets(w http.ResponseWriter, r *http.Request) {
	walletAddress := chi.URLParam(r, "walletAddress")
	if ok, _ := regexp.MatchString(`^0x[0-9a-fA-F]{40}$`, walletAddress); !ok {
		http.Error(w, "Invalid wallet address", http.StatusBadRequest)
		log.Println("Error at `handlerGetNumTickets`: Invalid wallet address")
		return
	}
	numTickets, err := cfg.DB.GetNumTickets(r.Context(), walletAddress)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithJSON(w, http.StatusOK, getNumTicketsResp{
				NumTickets: 0,
			})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error at `handlerGetNumTickets`: %s\n", err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, getNumTicketsResp{
		NumTickets: int(numTickets),
	})
}
