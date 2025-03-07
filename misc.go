package main

// Import Packages
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Builciber/blocknads-testmint-backend/internal/database"
	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type roleID string

type user struct {
	UserID   string `json:"id"`
	UserName string `json:"username"`
	Avatar   string `json:"avatar"`
}

type guildMember struct {
	User  user     `json:"user"`
	Roles []roleID `json:"roles"`
}

// The getUserGuildData() function is used to send an api
// request to the discord/users/@me/guilds/{guild.id}/member endpoint with
// the provided accessToken.
func (cfg *apiConfig) getUserGuildData(token string) (guildMember, error) {
	// Establish a new request object
	req, err := http.NewRequest("GET", fmt.Sprintf("https://discord.com/api/users/@me/guilds/%s/member", cfg.guildId), nil)

	// Handle the error
	if err != nil {
		return guildMember{}, err
	}
	// Set the request object's headers
	req.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{token},
	}
	// Send the http request
	client := http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	// Handle the error
	// If the response status isn't a success
	if resp.StatusCode != 200 || err != nil {
		// Read the http body
		body, _err := io.ReadAll(resp.Body)

		// Handle the read body error
		if _err != nil {
			return guildMember{}, _err
		}
		// Handle http response error
		return guildMember{},
			fmt.Errorf("status: %d, code: %v, body: %s",
				resp.StatusCode, err, string(body))
	}

	var data guildMember

	// Handle the error
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return guildMember{}, err
	}
	return data, nil
}

// The GetUserData() function is used to send an api
// request to the discord/users/@me endpoint with
// the provided accessToken.
func GetUserData(token string) (user, error) {
	// Establish a new request object
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)

	// Handle the error
	if err != nil {
		return user{}, err
	}
	// Set the request object's headers
	req.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{token},
	}
	// Send the http request
	client := http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	// Handle the error
	// If the response status isn't a success
	if resp.StatusCode != 200 || err != nil {
		// Read the http body
		body, _err := io.ReadAll(resp.Body)

		// Handle the read body error
		if _err != nil {
			return user{}, _err
		}
		// Handle http response error
		return user{},
			fmt.Errorf("status: %d, code: %v, body: %s",
				resp.StatusCode, err, string(body))
	}

	// Readable golang map used for storing
	// the response body
	var data user

	// Handle the error
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return user{}, err
	}
	return data, nil
}

func (cfg *apiConfig) writeNonceToDB(numNonces int) error {
	ok, err := cfg.DB.IsNonceColumnFilled(context.Background())
	if ok {
		return nil
	}
	if err != nil {
		return err
	}
	params := make([]database.CreateNoncesParams, numNonces)
	var nonce int16
	for i := 0; i < numNonces; i++ {
		id, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		currentTime := time.Now()
		params[i] = database.CreateNoncesParams{
			ID: pgtype.UUID{
				Bytes: id,
				Valid: true,
			},
			Nonce: nonce,
			CreatedAt: pgtype.Timestamp{
				Time:  currentTime,
				Valid: true,
			},
			UpdatedAt: pgtype.Timestamp{
				Time:  currentTime,
				Valid: true,
			},
		}
		nonce++
	}
	_, err = cfg.DB.CreateNonces(context.Background(), params)
	if err != nil {
		return err
	}
	return nil
}

func pollContractEvent(cfg *apiConfig, interval time.Duration) {
	tickChan := time.NewTicker(interval).C
	log.Println("Started contract event polling")
	w, err := web3.NewWeb3(cfg.rpcUrl)
	w.Eth.SetChainId(cfg.chainID)
	if err != nil {
		log.Fatal(err.Error())
	}
	startBlockasHexStr := "0x" + strconv.FormatUint(cfg.wlMintStartBlock, 16)
	toBlock := cfg.wlMintStartBlock + 9
	eventSignature := "Minted(address,uint256)"
	topicOne := crypto.Keccak256([]byte(eventSignature))
	for range tickChan {
		filter := types.Fliter{
			Address:   common.HexToAddress(cfg.contractAddress),
			FromBlock: startBlockasHexStr,
			ToBlock:   "0x" + strconv.FormatUint(toBlock, 16),
			Topics:    []string{common.Bytes2Hex(topicOne)},
		}
		events, err := w.Eth.GetLogs(&filter)
		if err != nil {
			continue
		}
		addresses := make([]string, len(events))
		for i, event := range events {
			params, err := w.Utils.DecodeParameters([]string{"address, uint256"}, []byte(event.Data))
			if err != nil {
				continue
			}
			minterAddress, ok := params[0].(string)
			if !ok {
				continue
			}
			addresses[i] = minterAddress
		}
	}
}
