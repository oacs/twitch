package main

import (
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
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

const (
	stateCallbackKey = "oauth-state-callback"
	oauthSessionName = "oauth-session"
	oauthTokenKey    = "oauth-token"
)

var (
	// Consider storing the secret in an environment variable or a dedicated storage system.
	scopes       = "channel:read:subscriptions"
	oauth2Config *clientcredentials.Config

	// TODO use a more secure cookieSecret
	cookieStore *sessions.CookieStore
)

func fetchApiToken() (string, error) {
	cookieSecret := os.Getenv("COOKIE_SECRET")
	cookieStore = sessions.NewCookieStore([]byte(cookieSecret))
	log.Debug("Fetching Twitch API token from ", TwitchTokenUrl)
	gob.Register(&oauth2.Token{})

	log.Debug("Fetching mock ")
	req, err := http.NewRequest("POST", TwitchTokenUrl, nil)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	query := req.URL.Query()
	query.Add("client_secret", clientSecret)
	query.Add("client_id", clientID)
	query.Add("grant_type", "client_credentials")
	query.Add("scopes", scopes)
	req.URL.RawQuery = query.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal("Error requesting mock info", err)
		return "", err
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body", err)
		return "", err
	}

	log.Debug("Response body: ", string(responseData))

	var authInfo struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(responseData, &authInfo)

	return authInfo.AccessToken, nil

	// 	oauth2Config = &clientcredentials.Config{
	// 		ClientID:     clientID,
	// 		ClientSecret: clientSecret,
	// 		TokenURL:     TwitchTokenUrl,
	// 	}

	// 	token, err := oauth2Config.Token(context.Background())
	// 	if err == nil {
	// 		log.Debug("Access token: %s\n", token.AccessToken)
	// 	} else {
	// 		log.Fatal(err)
	// 	}
	// 	return token.AccessToken, err
}

func fetchTwitchChannelInfo(token string) (channelInfo ChannelInfo, err error) {
	log.Debug("Fetching Twitch channel info\n")
	var req *http.Request
	if serverType == "mock" {
		req, err = http.NewRequest("GET", baseUrl+"/units/streams", nil)
	} else {
		req, err = http.NewRequest("GET", TwitchApiUrl+"/users?login=oacs69", nil)
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
