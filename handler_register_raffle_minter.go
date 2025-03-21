package main

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/chenzhijie/go-web3"
)

type registerRaffletMinterReq struct {
	WalletAddress string `json:"wallet_address"`
}

type registerRaffleMinterResp struct {
	Nonce     int16  `json:"nonce"`
	Signature string `json:"signature"`
}

func (cfg *apiConfig) handler_register_raffle_minter(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	reqBody := registerRaffletMinterReq{}
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ok, _ := regexp.MatchString(`^0x[0-9a-fA-F]{40}$`, reqBody.WalletAddress); !ok {
		http.Error(w, "Invalid wallet address", http.StatusBadRequest)
		return
	}
	minter, err := cfg.DB.GetRaffleMinter(r.Context(), reqBody.WalletAddress)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !minter.Nonce.Valid {
		http.Error(w, "ineligible user", http.StatusForbidden)
		return
	}
	wb, err := web3.NewWeb3(cfg.rpcUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = wb.Eth.SetAccount(cfg.signerPk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	msg, err := wb.Utils.EncodeParameters([]string{"uint256", "address", "uint256"}, []any{minter.Nonce, reqBody.WalletAddress, cfg.chainID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sig, err := wb.Eth.SignText(msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, registerRaffleMinterResp{
		Signature: hex.EncodeToString(sig),
		Nonce:     minter.Nonce.Int16,
	})
}
