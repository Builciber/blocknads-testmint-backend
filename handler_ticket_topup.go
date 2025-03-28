package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/Builciber/blocknads-testmint-backend/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ticketTopupResp struct {
	NewTotalTickets int `json:"new_total_tickets"`
}

func (cfg *apiConfig) handlerTicketTopup(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	reqBody := registerTicketPurchaseReq{}
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ok, _ := regexp.MatchString(`^0x[0-9a-fA-F]{40}$`, reqBody.WalletAddress); !ok {
		http.Error(w, "Invalid wallet address", http.StatusBadRequest)
		return
	}
	if reqBody.NumTickets == 0 {
		http.Error(w, "invalid number of tickets", http.StatusBadRequest)
		return
	}
	numTickets, err := cfg.DB.GetNumTickets(r.Context(), reqBody.WalletAddress)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Println("Error at `handlerTicketTopup`: Bad request")
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error at `handlerGetNumTickets`: %s\n", err.Error())
		return
	}
	if numTickets+int16(reqBody.NumTickets) > 100 {
		http.Error(w, "Total tickets purchased must be less than or equal to 100", http.StatusBadRequest)
		log.Println("Error at `handlerTicketTopup`: Total tickets purchased must be less than 100")
		return
	}
	err = cfg.DB.UpdateNumTickets(r.Context(), database.UpdateNumTicketsParams{
		WalletAddress: reqBody.WalletAddress,
		NumTickets:    int16(reqBody.NumTickets) + numTickets,
		UpdatedAt: pgtype.Timestamp{
			Time:  time.Now(),
			Valid: true,
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error at `handlerGetNumTickets`: %s\n", err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, ticketTopupResp{
		NewTotalTickets: int(numTickets) + int(reqBody.NumTickets),
	})
}
