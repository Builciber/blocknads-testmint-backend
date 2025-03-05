package main

// Import Packages
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
