package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

func createAuthenticationLink() string {
	url, err := url.Parse("https://discord.com/oauth2/authorize")
	if err != nil {
		log.Fatal(err)
	}

	query := url.Query()
	query.Set("response_type", "code")
	query.Set("client_id", DISCORD_CLIENT_ID)
	query.Set("scope", "guilds")
	query.Set("redirect_uri", DISCORD_REDIRECT_URL)

	url.RawQuery = query.Encode()

	return url.String()
}

func isUserInGuild(code string) bool {
	accessToken := getAccessTokenFromCode(code)
	doesGuildExist := doesGuildExist(accessToken, DISCORD_TARGET_GUILD_ID)

	revokeAccessToken(accessToken)
	return doesGuildExist
}

func getAccessTokenFromCode(code string) string {
	u := fmt.Sprintf(
		"https://%s:%s@discord.com/api/v10/oauth2/token",
		DISCORD_CLIENT_ID,
		DISCORD_CLIENT_SECRET)

	var resp struct {
		AccessToken string `json:"access_token"`
	}

	log.Println(u)

	res, err := httpClient.PostForm(u, url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {DISCORD_REDIRECT_URL},
	})

	if err != nil {
		log.Println(err.Error())
		return ""
	}

	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		log.Println(err.Error())
		return ""
	}

	return resp.AccessToken
}

func revokeAccessToken(accessToken string) {
	u := fmt.Sprintf(
		"https://%s:%s@discord.com/api/v10/oauth2/token/revoke",
		DISCORD_CLIENT_ID,
		DISCORD_CLIENT_SECRET)

	httpClient.PostForm(u, url.Values{
		"data":            {accessToken},
		"token_type_hint": {"access_token"},
	})
}

func doesGuildExist(accessToken, guildId string) bool {
	var resp []struct {
		Id string `json:"id"`
	}

	req, err := http.NewRequest("GET", "https://discord.com/api/v10/users/@me/guilds", nil)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err := httpClient.Do(req)

	if err != nil {
		log.Println(err.Error())
		return false
	}

	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	for _, guild := range resp {
		if guild.Id == guildId {
			return true
		}
	}

	return false
}
