package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Builciber/blocknads-testmint-backend/internal/auth"
	"github.com/Builciber/blocknads-testmint-backend/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/realTristan/disgoauth"
)

func (cfg *apiConfig) handler_auth_callback(dc *disgoauth.Client) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		nonce := r.URL.Query().Get("state")
		cfg.mut.Lock()
		ok := cfg.oauthStates[nonce]
		if !ok {
			http.Redirect(w, r, fmt.Sprintf("%s?status=failed&reason=%s", cfg.clientCallbackURL, url.QueryEscape("unknown state parameter")), http.StatusFound)
			return
		}
		delete(cfg.oauthStates, nonce)
		cfg.mut.Unlock()
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Redirect(w, r, fmt.Sprintf("%s?status=failed&reason=%s", cfg.clientCallbackURL, url.QueryEscape("authentication failed")), http.StatusFound)
			return
		}
		accessToken, err := dc.GetOnlyAccessToken(code)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s?status=failed&reason=%s", cfg.clientCallbackURL, url.QueryEscape("internal server error")), http.StatusFound)
			return
		}
		user, err := GetUserData(accessToken)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s?status=failed&reason=%s", cfg.clientCallbackURL, url.QueryEscape("internal server error")), http.StatusFound)
			return
		}
		guildMemberData, err := cfg.getUserGuildData(accessToken)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s?status=failed&reason=%s", cfg.clientCallbackURL, url.QueryEscape("internal server error")), http.StatusFound)
			return
		}
		roles := guildMemberData.Roles
		ok = false
		for _, role := range roles {
			if role == roleID(cfg.verfiedRoleId) {
				ok = true
				break
			}
		}
		if !ok {
			http.Redirect(w, r, fmt.Sprintf("%s?status=failed&reason=%s", cfg.clientCallbackURL, url.QueryEscape("ineligible user")), http.StatusFound)
			return
		}
		signedSessionToken, err := auth.CreateJWT(user.UserID, cfg.sessionSecret)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s?status=failed&reason=%s", cfg.clientCallbackURL, url.QueryEscape("internal server error")), http.StatusFound)
			return
		}
		sessionCookie := http.Cookie{
			Name:       "mint-session",
			Value:      signedSessionToken,
			Expires:    time.Now().UTC().Add(4 * time.Hour),
			Domain:     cfg.domain,
			Path:       "/",
			HttpOnly:   true,
			Secure:     false,
			SameSite:   http.SameSiteLaxMode,
			RawExpires: time.Now().UTC().Add(4 * time.Hour).String(),
		}
		w.Header().Add("Set-Cookie", sessionCookie.String())
		err = cfg.DB.UpdateWhitelistMinterAfterAuth(r.Context(), database.UpdateWhitelistMinterAfterAuthParams{
			DiscordID: user.UserID,
			DiscordUsername: pgtype.Text{
				String: user.UserName,
				Valid:  true,
			},
			AvatarHash: pgtype.Text{
				String: user.Avatar,
				Valid:  true,
			},
			UpdatedAt: pgtype.Timestamp{
				Time:  time.Now(),
				Valid: true,
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("%s?status=success&username=%s&avatar=%s", cfg.clientCallbackURL, user.UserName, user.Avatar), http.StatusFound)
	}
	return http.HandlerFunc(fn)
}
