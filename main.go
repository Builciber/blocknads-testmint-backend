package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/Builciber/blocknads-testmint-backend/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/realTristan/disgoauth"
)

type apiConfig struct {
	chainID           int64
	wlMintStartBlock  uint64
	sessionSecret     string
	domain            string
	signerPk          string
	verfiedRoleId     string
	guildId           string
	clientID          string
	clientSecret      string
	clientCallbackURL string
	clientOrigin      string
	rpcUrl            string
	contractAddress   string
	ownerPK           string
	DB                *database.Queries
	mut               *sync.RWMutex
	oauthStates       map[string]bool
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("CONN")
	sessionSecret := os.Getenv("SESSION_SECRET")
	domain := os.Getenv("DOMAIN")
	signerPk := os.Getenv("SIGNER_PK")
	verfiedRoleId := os.Getenv("VERIFIED_ROLE_ID")
	clientCallbackURL := os.Getenv("CLIENT_CALLBACK_URL")
	clientOrigin := os.Getenv("CLIENT_ORIGIN")
	guildId := os.Getenv("GUILD_ID")
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	rpcUrl := os.Getenv("RPC_URL")
	contractAddress := os.Getenv("CONTRACT_ADDRESS")
	ownerPK := os.Getenv("OWNER_PK")
	chainID, err := strconv.Atoi(os.Getenv("CHAIN_ID"))
	if err != nil {
		log.Fatal(err.Error())
	}
	wlMintStartBlock, err := strconv.ParseUint(os.Getenv("WL_MINT_START_BLOCK"), 10, 0)
	if err != nil {
		log.Fatal(err.Error())
	}
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	dbQueries := database.New(db)
	apiMux := chi.NewRouter()
	cfg := apiConfig{
		chainID:           int64(chainID),
		DB:                dbQueries,
		sessionSecret:     sessionSecret,
		mut:               &sync.RWMutex{},
		oauthStates:       make(map[string]bool),
		domain:            domain,
		signerPk:          signerPk,
		verfiedRoleId:     verfiedRoleId,
		guildId:           guildId,
		clientID:          clientId,
		clientSecret:      clientSecret,
		clientCallbackURL: clientCallbackURL,
		clientOrigin:      clientOrigin,
		rpcUrl:            rpcUrl,
		ownerPK:           ownerPK,
		wlMintStartBlock:  wlMintStartBlock,
		contractAddress:   contractAddress,
	}
	err = cfg.writeNonceToDB(99)
	if err != nil {
		log.Fatal(err)
	}
	apiMux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.clientOrigin},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	var dc *disgoauth.Client = disgoauth.Init(&disgoauth.Client{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURI:  fmt.Sprintf("https://%s/api/auth/callback", cfg.domain),
		Scopes:       []string{disgoauth.ScopeIdentify, "guilds.members.read"},
	})
	apiMux.Get("/auth", cfg.handler_auth(dc))
	apiMux.Get("/auth/callback", cfg.handler_auth_callback(dc))
	apiMux.Get("/auth/logout", cfg.handler_logout)
	apiMux.Post("/register/raffle_minter", cfg.handler_register_raffle_minter)
	apiMux.Post("/register/ticket_purchase", cfg.handler_register_ticket_purchase)
	apiMux.Post("/register/whitelist_minter", cfg.handler_register_whitelist_minter)
	//apiMux.Post("/test/whitelistMint", cfg.handlerWhitelistMintTest)
	//apiMux.Get("/test/issueSessionToken", cfg.handlerIssueSessionToken())
	apiMux.Mount("/api/", apiMux)
	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: apiMux,
	}
	log.Println("Starting server on localhost at port 8080")
	err = server.ListenAndServe()
	log.Fatal(err)
}
