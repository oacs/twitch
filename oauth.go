package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
)

type ChannelInfo struct {
	DisplayName     string `json:"display_name"`
	Login           string `json:"login"`
	Type            string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	ProfileImageURL string `json:"profile_image_url"`
	OfflineImageURL string `json:"offline_image_url"`
	ID              string `json:"id"`
}

// HumanReadableError represents error information
// that can be fed back to a human user.
// This prevents internal state that might be sensitive
// being leaked to the outside world.
type HumanReadableError interface {
	HumanError() string
	HTTPCode() int
}

// HumanReadableWrapper implements HumanReadableError
type HumanReadableWrapper struct {
	ToHuman string
	Code    int
	error
}

type Handler func(http.ResponseWriter, *http.Request) error

const (
	stateCallbackKey = "oauth-state-callback"
	oauthSessionName = "oauth-session"
	oauthTokenKey    = "oauth-token"
)

var (
	// Consider storing the secret in an environment variable or a dedicated storage system.
	scopes       = []string{"user:read:email"}
	oauth2Config *clientcredentials.Config

	// TODO use a more secure cookieSecret
	cookieSecret = []byte("Please use a more sensible secret than this one")
	cookieStore  = sessions.NewCookieStore(cookieSecret)
)

func (h HumanReadableWrapper) HumanError() string { return h.ToHuman }
func (h HumanReadableWrapper) HTTPCode() int      { return h.Code }

// HandleLogin is a Handler that redirects the user to Twitch for login, and provides the 'state'
// parameter which protects against login CSRF.

func callWebhook(token string, channelId string) (err error) {
	// use the token to get a user's profile and email address, and store both in a session.
	// TODO ngroks for receiving the values
	// TODO use a more secure cookieSecret

	var jsonStr = []byte(fmt.Sprintf(`{
    "type": "channel.follow",
    "version": "1",
    "condition": {
        "broadcaster_user_id": "%s"
    },
    "transport": {
        "method": "webhook",
        "callback": "https://8c75-212-201-41-171.ngrok.io/callback",
        "secret": "s3cRe72345679"
    }}`, channelId))

	req, err := http.NewRequest("GET", TwitchApiUrl+"/eventsub/subscriptions", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Client-ID", clientID)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal("Error calling API twitch channel info", err)
		return err
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body", err)
		return err
	}

	var response struct {
		Data []struct {
			Id string `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(responseData, &response)

	for _, v := range response.Data {
		fmt.Printf("v.Id: %v\n", v.Id)

		req, err = http.NewRequest("DELETE", TwitchApiUrl+"/eventsub/subscriptions?id="+v.Id, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Client-ID", clientID)
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Fatal("Error calling API twitch channel info", err)
			return err
		}

		responseData, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading response body", err)
			return err
		}
	}
	// log.Debug(string(responseData))
	return nil
}

func fetchTwitchChannelInfo(token string) (channelInfo ChannelInfo, err error) {
	log.Debug("Fetching Twitch channel info\n")
	var req *http.Request
	if serverType == "mock" {
		req, err = http.NewRequest("GET", baseUrl+"/units/streams", nil)
	} else {
		req, err = http.NewRequest("GET", TwitchApiUrl+"/users?login=twitchdev", nil)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Client-ID", clientID)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal("Error calling API twitch channel info", err)
		return channelInfo, err
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body", err)
		return channelInfo, err
	}

	log.Debug("Response from channel info ", string(responseData))

	var channelArrayInfo struct {
		Data []ChannelInfo `json:"data"`
	}
	json.Unmarshal(responseData, &channelArrayInfo)

	return channelArrayInfo.Data[0], nil
}

// AnnotateError wraps an error with a message that is intended for a human end-user to read,
// plus an associated HTTP error code.
func AnnotateError(err error, annotation string, code int) error {
	if err == nil {
		return nil
	}
	return HumanReadableWrapper{ToHuman: annotation, error: err}
}

func middleware(h Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		// parse POST body, limit request size
		if err = r.ParseForm(); err != nil {
			return AnnotateError(err, "Something went wrong! Please try again.", http.StatusBadRequest)
		}

		return h(w, r)
	}
}

// errorHandling is a middleware that centralises error handling.
// this prevents a lot of duplication and prevents issues where a missing
// return causes an error to be printed, but functionality to otherwise continue
// see https://blog.golang.org/error-handling-and-go
func errorHandling(handler func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			var errorString string = "Something went wrong! Please try again."
			var errorCode int = 500

			if v, ok := err.(HumanReadableError); ok {
				errorString, errorCode = v.HumanError(), v.HTTPCode()
			}

			log.Fatal(err)
			w.Write([]byte(errorString))
			w.WriteHeader(errorCode)
			return
		}
	})
}

func handleFunc(path string, handler Handler) {
	http.Handle(path, errorHandling(middleware(handler)))
}
