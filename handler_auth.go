package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/Builciber/blocknads-testmint-backend/internal/auth"
	"github.com/realTristan/disgoauth"
)

type discordAuthResp struct {
	DiscordID string `json:"discord_id"`
	Avatar    string `json:"avatar"`
}

func (cfg *apiConfig) handler_auth(dc *disgoauth.Client) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("mint-session")
		if err == nil {
			discordID, err := auth.ValidateJWT(cookie.Value, cfg.sessionSecret)
			if err != nil {
				if err.Error() == "session is invalid or expired" {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondWithJSON(w, http.StatusAccepted, discordAuthResp{DiscordID: discordID})
			return
		}
		//If cookie was not found, we call Discord's authentication endpoint
		b := make([]byte, 32)
		rand.Read(b)
		state := base64.URLEncoding.EncodeToString(b)
		cfg.mut.Lock()
		cfg.oauthStates[state] = true
		cfg.mut.Unlock()
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		dc.RedirectHandler(w, r, state)
	}

	return http.HandlerFunc(fn)
}
