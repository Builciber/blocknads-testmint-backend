package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

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
	rafflePeriodStart uint64
	DB                *database.Queries
	dbConn            *pgxpool.Pool
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
	rafflePeriodStart, err := strconv.ParseUint(os.Getenv("RAFFLE_PERIOD_START"), 10, 0)
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
		dbConn:            db,
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
		rafflePeriodStart: rafflePeriodStart,
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
	clientMap := make(map[*client]struct{})
	clientMu := &sync.Mutex{}
	raffleStartChan := make(chan uint64, 1)
	apiMux.Get("/auth", cfg.handlerAuth(dc))
	apiMux.Get("/auth/callback", cfg.handlerAuthCallback(dc))
	apiMux.Get("/auth/logout", cfg.handlerLogout)
	apiMux.Get("/tickets/bought/{walletAddress}", cfg.handlerGetNumTickets)
	apiMux.Get("/raffle/check/{walletAddress}", cfg.handlerCheckRaffleWinner)
	apiMux.Get("/stream/events", cfg.handlerStreamEvents(clientMap, clientMu))
	apiMux.Post("/register/raffle_minter", cfg.handlerRegisterRaffleMinter)
	apiMux.Post("/register/ticket_purchase", cfg.handlerRegisterTicketPurchase)
	apiMux.Post("/register/whitelist_minter", cfg.handlerRegisterWhitelistMinter)
	apiMux.Post("/tickets/topup", cfg.handlerTicketTopup)
	//apiMux.Post("/test/whitelistMint", cfg.handlerWhitelistMintTest)
	//apiMux.Get("/test/issueSessionToken", cfg.handlerIssueSessionToken())
	//apiMux.Get("/test", cfg.quickTest)
	//apiMux.Post("/genFakeTicketBuyers", cfg.generateFakeTicketBuyers)
	apiMux.Mount("/api/", apiMux)
	hasRaffled, err := cfg.DB.GetRaffleState(context.Background())
	if err != nil {
		log.Fatal(err.Error())
	}
	if !hasRaffled {
		go listenForRaffleStart(&cfg, raffleStartChan, &hasRaffled)
	}
	go pollBlockNumber(&cfg, clientMap, clientMu, 10*time.Second, raffleStartChan, &hasRaffled)
	go pollMintedEvent(&cfg, clientMap, clientMu, 10*time.Second)
	log.Println("Started server on localhost at port 8080")
	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: apiMux,
	}
	err = server.ListenAndServe()
	log.Fatal(err)
}
