package main

import (
	"net/http"
	"time"

	"github.com/Builciber/blocknads-testmint-backend/internal/auth"
	"github.com/realTristan/disgoauth"
)

func (cfg *apiConfig) handler_auth_callback(dc *disgoauth.Client) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		nonce := r.URL.Query().Get("state")
		cfg.mut.Lock()
		ok := cfg.oauthStates[nonce]
		if !ok {
			http.Error(w, "click-jacking suspected", http.StatusUnauthorized)
			return
		}
		delete(cfg.oauthStates, nonce)
		cfg.mut.Unlock()
		code := r.URL.Query().Get("code")
		accessToken, err := dc.GetOnlyAccessToken(code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		userData, err := GetUserData(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, ok := userData["id"].(string)
		if !ok {
			http.Error(w, "unexpected ID type", http.StatusInternalServerError)
			return
		}
		userGuildData, err := cfg.getUserGuildData(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		roles, ok := userGuildData["roles"].([]string)
		if !ok {
			http.Error(w, "roles information absent", http.StatusInternalServerError)
			return
		}
		ok = false
		for _, e := range roles {
			if e == cfg.verfiedRoleId {
				ok = true
				break
			}
		}
		if !ok {
			http.Error(w, "ineligible user", http.StatusUnauthorized)
			return
		}
		signedSessionToken, err := auth.CreateJWT(id, cfg.sessionSecret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sessionCookie := http.Cookie{
			Name:       "mint-session",
			Value:      signedSessionToken,
			Expires:    time.Now().UTC().Add(4 * time.Hour),
			Domain:     cfg.domain,
			Path:       "/auth/",
			HttpOnly:   true,
			Secure:     false,
			SameSite:   http.SameSiteLaxMode,
			RawExpires: time.Now().UTC().Add(4 * time.Hour).String(),
		}
		w.Header().Add("Set-Cookie", sessionCookie.String())
		respondWithJSON(w, http.StatusOK, discordAuthResp{DiscordID: id})
	}
	return http.HandlerFunc(fn)
}
