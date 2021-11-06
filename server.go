package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"time"

	// "encoding/json"
	"flag"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const ()

var (
	serverType     string
	clientSecret   string
	clientID       string
	TwitchTokenUrl string
	TwitchApiUrl   string
	userID         string
	baseUrl        string
	httpClient     = &http.Client{}
)

func init() {
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	// LOG_LEVEL not set, let's default to debug
	if !ok {
		lvl = "debug"
	}
	// parse string, this is built-in feature of logrus
	ll, err := log.ParseLevel(lvl)
	if err != nil {
		ll = log.DebugLevel
	}
	// set global log level
	log.SetLevel(ll)
}

func startMockServer() {
	log.Info("Starting mock server")
	// TODO fetch this from the mock server
	clientSecret = os.Getenv("TWITCH_CLIENT_SECRET_MOCKED")
	clientID = os.Getenv("TWITCH_CLIENT_ID_MOCKED")
	userID = os.Getenv("TWITCH_USER_ID_MOCKED")
	baseUrl = "http://localhost:8080"
	TwitchTokenUrl = baseUrl + "/auth/token"
	TwitchApiUrl = baseUrl + "/mock"
}
func startTwitchServer() {
	log.Info("Starting twitch server")
	clientSecret = os.Getenv("TWITCH_CLIENT_SECRET")
	clientID = os.Getenv("TWITCH_CLIENT_ID")
	TwitchTokenUrl = "https://id.twitch.tv/oauth2/token"
	TwitchApiUrl = "https://api.twitch.tv/helix"
}
func loadEnvVariables() (err error) {
	log.Debug("Fetching .env ")
	err = godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Error loading .env file")
		return err
	}
	return
}

func fetchApiToken() (string, error) {
	log.Debug("Fetching Twitch API token from ", TwitchTokenUrl)
	gob.Register(&oauth2.Token{})

	if serverType == "mock" {
		log.Debug("Fetching mock ")
		req, err := http.NewRequest("POST", TwitchTokenUrl, nil)
		if err != nil {
			log.Fatal(err)
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		query := req.URL.Query()
		query.Add("client_secret", clientSecret)
		query.Add("grant_type", "client_credentials")
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
	}

	oauth2Config = &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     TwitchTokenUrl,
	}

	token, err := oauth2Config.Token(context.Background())
	if err == nil {
		log.Debug("Access token: %s\n", token.AccessToken)
	} else {
		log.Fatal(err)
	}
	return token.AccessToken, err
}

type SubscriptionReq struct {
	Subscription struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Type      string `json:"type"`
		Version   string `json:"version"`
		Condition struct {
			BroadcasterUserID string `json:"broadcaster_user_id"`
		} `json:"condition"`
		Transport struct {
			Method   string `json:"method"`
			Callback string `json:"callback"`
		} `json:"transport"`
		CreatedAt time.Time `json:"created_at"`
		Cost      int       `json:"cost"`
	} `json:"subscription"`
	Event struct {
		UserID               string `json:"user_id"`
		UserLogin            string `json:"user_login"`
		UserName             string `json:"user_name"`
		BroadcasterUserID    string `json:"broadcaster_user_id"`
		BroadcasterUserLogin string `json:"broadcaster_user_login"`
		BroadcasterUserName  string `json:"broadcaster_user_name"`
		Tier                 string `json:"tier"`
		IsGift               bool   `json:"is_gift"`
	} `json:"event"`
}

func callback(w http.ResponseWriter, r *http.Request) (err error) {
	log.Debug("Callback called")

	requestData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal("Error reading response body", err)
		return err
	}
	var requestParsed SubscriptionReq
	err = json.Unmarshal(requestData, &requestParsed)
	if err != nil {
		log.Fatal("Error parsing request body", err)
		return err
	}

	for _, con := range c {
		con.WriteJSON(requestParsed)
	}
	return
}
func main() {
	// Go encoding for gorilla/sessions
	err := loadEnvVariables()
	if err != nil {
		return
	}

	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 {
		serverType = argsWithoutProg[0]
	} else {
		serverType = "twitch"
	}
	// TODO
	switch serverType {
	case "mock":
		startMockServer()
	case "twitch":
		startTwitchServer()
	}

	// Fetch API Token
	token, err := fetchApiToken()
	if err != nil {
		log.Fatal("Error fetching API token")
		return
	}

	channelInfo, err := fetchTwitchChannelInfo(token)
	if err != nil {
		log.Fatal("Error fetching Twitch Channel Info")
		return
	}

	go callWebhook(token, channelInfo.ID)
	twitchIRCConnect()
	go callWebhook(token, channelInfo.ID)
	// Websocket init
	log.Debug("Creating static dir")
	fs := http.FileServer(http.Dir("./static"))
	flag.Parse()
	log.Info("Serving static files")
	http.Handle("/", fs)
	handleFunc("/chat", chat)
	handleFunc("/callback", callback)

	log.Info("Started running on http://localhost:7001")
	log.Info(http.ListenAndServe(":7001", nil))
}
