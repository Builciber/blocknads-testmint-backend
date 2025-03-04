package main

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/Builciber/blocknads-testmint-backend/internal/auth"
	"github.com/chenzhijie/go-web3/utils"
	"github.com/ethereum/go-ethereum/crypto"
)

type registerWhitelistMintersReq struct {
	WalletAddress string `json:"wallet_address"`
}

type registerWhitelistMintersResp struct {
	DiscordID uint64 `json:"discord_id"`
	Nonce     int16  `json:"nonce"`
	Signature string `json:"signature"`
}

func (cfg *apiConfig) handler_register_whitelist_minter(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("mint-session")
	if err != nil {
		http.Error(w, "session is missing", http.StatusUnauthorized)
		return
	}
	discordID, err := auth.ValidateJWT(cookie.Value, cfg.sessionSecret)
	if err != nil {
		if err.Error() == "session is invalid or expired" {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	reqBody := registerWhitelistMintersReq{}
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ok, _ := regexp.MatchString(`^0x[0-9a-fA-F]{40}$`, reqBody.WalletAddress); !ok {
		http.Error(w, "Invalid wallet address", http.StatusBadRequest)
		return
	}
	minter, err := cfg.DB.GetWhitelistMinterById(r.Context(), discordID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pk, err := crypto.HexToECDSA(cfg.signerPk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	idAsUint, err := strconv.ParseUint(discordID, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	util := utils.Utils{}
	msg, err := util.EncodeParameters([]string{"uint256", "uint64", "address", "uint256"}, []any{minter.Nonce, idAsUint, reqBody.WalletAddress, cfg.chainID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	msgHash := crypto.Keccak256(msg)
	sig, err := crypto.Sign(msgHash, pk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, registerWhitelistMintersResp{
		Signature: hex.EncodeToString(sig),
		DiscordID: idAsUint,
		Nonce:     minter.Nonce,
	})
}
