package main

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/Builciber/blocknads-testmint-backend/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type registerTicketPurchaseReq struct {
	WalletAddress string `json:"wallet_address"`
	NumTickets    uint8  `json:"num_tickets"`
}

func (cfg *apiConfig) handler_register_ticket_purchase(w http.ResponseWriter, r *http.Request) {
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
	err = cfg.DB.CreateTicketBuyer(r.Context(), database.CreateTicketBuyerParams{
		WalletAddress: reqBody.WalletAddress,
		NumTickets:    int16(reqBody.NumTickets),
		CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
