package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerLogout(w http.ResponseWriter, r *http.Request) {
	sessionCookie := http.Cookie{
		Name:     "mint-session",
		MaxAge:   -1,
		Domain:   cfg.cookieDomain,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	w.Header().Add("Set-Cookie", sessionCookie.String())
	http.Redirect(w, r, fmt.Sprintf("%s?status=success", cfg.clientCallbackURL), http.StatusFound)
}
