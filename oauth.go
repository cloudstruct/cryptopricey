// Taken from https://raw.githubusercontent.com/demisto/slack/master/examples/oauth/oauth.go
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/demisto/slack"

	"golang.org/x/oauth2"
)

type state struct {
	auth string
	ts   time.Time
}

// globalState is an example of how to store a state between calls
var globalState state

func logerr(n int, err error) {
	if err != nil {
		log.Printf("Write failed: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	logerr(w.Write([]byte(err)))
}

// addToSlack initializes the oauth process and redirects to Slack
func addToSlack(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")

	// Just generate random state
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		writeError(w, 500, err.Error())
	}
	globalState = state{auth: hex.EncodeToString(b), ts: time.Now()}
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"client"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://slack.com/oauth/authorize",
			TokenURL: "https://slack.com/api/oauth.access", // not actually used here
		},
	}
	url := conf.AuthCodeURL(globalState.auth)
	http.Redirect(w, r, url, http.StatusFound)
}

// auth receives the callback from Slack, validates and displays the user information
func auth(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")

	state := r.FormValue("state")
	code := r.FormValue("code")
	errStr := r.FormValue("error")
	if errStr != "" {
		writeError(w, 401, errStr)
		return
	}
	if state == "" || code == "" {
		writeError(w, 400, "Missing state or code")
		return
	}
	if state != globalState.auth {
		writeError(w, 403, "State does not match")
		return
	}
	// As an example, we allow only 5 min between requests
	if time.Since(globalState.ts) > 5*time.Minute {
		writeError(w, 403, "State is too old")
		return
	}
	token, err := slack.OAuthAccess(clientID, clientSecret, code, "")
	if err != nil {
		writeError(w, 401, err.Error())
		return
	}
	s, err := slack.New(slack.SetToken(token.AccessToken))
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	// Get our own user id
	test, err := s.AuthTest()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	logerr(w.Write([]byte(fmt.Sprintf("OAuth successful for team %s and user %s", test.Team, test.User))))
}

// home displays the add-to-slack button
func home(w http.ResponseWriter, r *http.Request) {
	logerr(w.Write([]byte(`<html><head><title>Slack OAuth Test</title></head><body><a href="/add">Add To Slack</a></body></html>`)))
}
