package main

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"strconv"

	"github.com/Builciber/blocknads-testmint-backend/internal/auth"
	"github.com/Builciber/blocknads-testmint-backend/internal/database"
	"github.com/chenzhijie/go-web3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v5/pgtype"
)

type registerWhitelistMintersReq struct {
	WalletAddress string `json:"wallet_address"`
}

type registerWhitelistMintersResp struct {
	DiscordID uint64 `json:"id"`
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
	minter, err := cfg.DB.GetWhitelistMinterById(r.Context(), pgtype.Text{String: discordID, Valid: true})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !(!minter.WalletAddress.Valid || minter.WalletAddress.String == reqBody.WalletAddress) {
		http.Error(w, "only one wallet per discord user", http.StatusNotAcceptable)
		return
	}
	idAsUint, err := strconv.ParseUint(discordID, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	wb, err := web3.NewWeb3(cfg.rpcUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	wb.Eth.SetChainId(cfg.chainID)
	err = wb.Eth.SetAccount(cfg.signerPk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	msg, err := wb.Utils.EncodeParameters([]string{"uint256", "uint64", "address", "uint256"}, []any{big.NewInt(int64(minter.Nonce)), idAsUint, common.HexToAddress(reqBody.WalletAddress), big.NewInt(cfg.chainID)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	msgHash := crypto.Keccak256(msg)
	sig, err := wb.Eth.SignText(msgHash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !minter.WalletAddress.Valid {
		err = cfg.DB.AddWhitelistMintWallet(r.Context(), database.AddWhitelistMintWalletParams{
			DiscordID: pgtype.Text{
				String: discordID,
				Valid:  true,
			},
			WalletAddress: pgtype.Text{
				String: reqBody.WalletAddress,
				Valid:  true,
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	respondWithJSON(w, http.StatusOK, registerWhitelistMintersResp{
		Signature: hex.EncodeToString(sig),
		DiscordID: idAsUint,
		Nonce:     minter.Nonce,
	})
}
