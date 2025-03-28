package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/Builciber/blocknads-testmint-backend/internal/auth"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/realTristan/disgoauth"
)

func (cfg *apiConfig) handlerAuth(dc *disgoauth.Client) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("mint-session")
		if err == nil {
			discordID, err := auth.ValidateJWT(cookie.Value, cfg.sessionSecret)
			if err != nil {
				if err.Error() == "session is invalid or expired" {
					http.Redirect(w, r, fmt.Sprintf("%s?status=failed&reason=%s", cfg.clientCallbackURL, "unauthorized"), http.StatusFound)
					log.Println(err.Error())
					return
				}
				http.Redirect(w, r, fmt.Sprintf("%s?status=failed&reason=%s", cfg.clientCallbackURL, url.QueryEscape("internal server error")), http.StatusFound)
				log.Println(err.Error())
				return
			}
			minter, err := cfg.DB.GetWhitelistMinterById(r.Context(), pgtype.Text{String: discordID, Valid: true})
			if err != nil {
				http.Redirect(w, r, fmt.Sprintf("%s?status=failed&reason=%s", cfg.clientCallbackURL, url.QueryEscape("internal server error")), http.StatusFound)
				log.Println(err.Error())
				return
			}
			http.Redirect(w, r, fmt.Sprintf("%s?status=success&username=%s&avatar=%s&userid=%s", cfg.clientCallbackURL, minter.DiscordUsername.String, minter.AvatarHash.String, minter.DiscordID.String), http.StatusFound)
			return
		}
		//If cookie was not found, we call Discord's authentication endpoint
		b := make([]byte, 32)
		rand.Read(b)
		state := base64.URLEncoding.EncodeToString(b)
		cfg.mut.Lock()
		cfg.oauthStates[state] = true
		cfg.mut.Unlock()
		dc.RedirectHandler(w, r, state)
	}

	return http.HandlerFunc(fn)
}
