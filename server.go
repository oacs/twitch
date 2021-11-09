package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
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
	token          string
	Port           string
	ngrokUrl       string
	channelInfo    ChannelInfo
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
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return err
	}
	return
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

func main() {
	err := loadEnvVariables()
	Port = os.Getenv("PORT")
	if err != nil {
		return
	}
	ngrokUrl, err = runNgrok()
	if err != nil {
		log.Fatal("Error running ngrok", err)
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
	token, err = fetchApiToken()
	if err != nil {
		log.Fatal("Error fetching API token")
		return
	}

	channelInfo, err = fetchTwitchChannelInfo(token)
	if err != nil {
		log.Fatal("Error fetching Twitch Channel Info")
		return
	}

	twitchIRCConnect()
	// Websocket init
	log.Debug("Creating static dir")
	fs := http.FileServer(http.Dir("./static"))
	flag.Parse()
	log.Info("Serving static files")
	http.Handle("/", fs)
	handleFunc("/chat", chat)
	handleFunc("/callback", responseChallengeCallback)
	handleFunc("/create", create)

	log.Info("Started running on http://localhost:" + Port)
	log.Info(http.ListenAndServe(Port, nil))
}
