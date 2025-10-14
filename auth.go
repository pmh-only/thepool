package main

import (
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func isVaildToken(tokenStr string) bool {
	if !DISCORD_AUTH_ENABLED {
		return true
	}

	tokenSlice := strings.Split(tokenStr, " ")
	if len(tokenSlice) != 2 {
		return false
	}

	if tokenSlice[0] != "Bearer" {
		return false
	}

	_, err := jwt.Parse(
		tokenSlice[1],
		func(t *jwt.Token) (any, error) {
			return []byte(DISCORD_AUTH_TOKEN), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer("thepool. (hey, stop. plz do not break this thing oka? :)"),
		jwt.WithExpirationRequired())

	return err == nil
}

func getVaildToken() string {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": "thepool. (hey, stop. plz do not break this thing oka? :)",
			"exp": time.Now().Unix() + 5*60*60,
		},
	)

	tokenStr, err := token.SignedString([]byte(DISCORD_AUTH_TOKEN))
	if err != nil {
		log.Fatal(err)
	}

	return tokenStr
}
