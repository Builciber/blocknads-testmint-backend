package main

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/Builciber/blocknads-testmint-backend/internal/database"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type registerTicketPurchaseReq struct {
	WalletAddress string `json:"wallet_address"`
	NumTickets    uint8  `json:"num_tickets"`
}

func (cfg *apiConfig) handlerRegisterTicketPurchase(w http.ResponseWriter, r *http.Request) {
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
	if reqBody.NumTickets < 1 || reqBody.NumTickets > 100 {
		http.Error(w, "invalid number of tickets", http.StatusBadRequest)
		return
	}
	currentTime := time.Now()
	err = cfg.DB.CreateTicketBuyer(r.Context(), database.CreateTicketBuyerParams{
		WalletAddress: reqBody.WalletAddress,
		NumTickets:    int16(reqBody.NumTickets),
		CreatedAt:     pgtype.Timestamp{Time: currentTime, Valid: true},
		UpdatedAt:     pgtype.Timestamp{Time: currentTime, Valid: true},
	})
	if pgErr, ok := err.(*pgconn.PgError); ok {
		if pgErr.Code == "23505" {
			http.Error(w, "user has purchased tickets before", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
