package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

type Event struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

type Transport struct {
	Method   string `json:"method"`
	Callback string `json:"callback"`
	Secret   string `json:"secret"`
}

type Condition struct {
	BroadcasterUserId string `json:"broadcaster_user_id"`
}

type Subscription struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Cost      int       `json:"cost"`
	Condition Condition `json:"condition"`
	Transport Transport `json:"transport"`
	CreatedAt time.Time `json:"created_at"`
}

type SubscriptionCallback struct {
	Challenge    string       `json:"challenge"`
	Subscription Subscription `json:"subscription"`
	Event        Event        `json:"event"`
}

// delete this c1cdbfac-7103-4bb2-b1cd-5792e906a75b
type SubsciptionCreatedResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Cost      int       `json:"cost"`
	Condition Condition `json:"condition"`
	Transport Transport `json:"transport"`
	CreatedAt time.Time `json:"created_at"`
}

func printNgrok(scanner *bufio.Reader) {
	for {
		line, err := scanner.ReadString('\n')
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Debug(line)
	}
}

func runNgrok() (ngrok string, err error) {
	// create a shell to run ngrok and get the url
	cmd := exec.Command("./ngrok", "http", "7001", "--log", "stdout")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
		return "", err
	} else {
		log.Debug("Ngrok started")
	}
	var ngrokTunnelRegex = regexp.MustCompile(`https:\/\/.*\.io`)
	var ngrokUrl string
	// create stdout reader
	scanner := bufio.NewReader(stdout)
	for ngrokUrl == "" {
		line, err := scanner.ReadString('\n')
		if err != nil {
			log.Fatal(err)
			return "", err
		}
		if ngrokTunnelRegex.MatchString(line) {
			ngrokUrl = ngrokTunnelRegex.FindString(line)
			break
		} else {
			log.Debug(line)
		}

	}
	go printNgrok(scanner)

	log.Debug("Ngrok url: ", ngrokUrl)
	return ngrokUrl, nil
}

func create(w http.ResponseWriter, r *http.Request) (err error) {
	removeAllSubs(token, channelInfo.ID)
	// Fetch API Token
	cookieSecret := os.Getenv("COOKIE_SECRET")
	createSubscription("channel.follow", token, channelInfo.ID, ngrokUrl, cookieSecret)
	createSubscription("channel.update", token, channelInfo.ID, ngrokUrl, cookieSecret)
	createSubscription("channel.subscribe", token, channelInfo.ID, ngrokUrl, cookieSecret)

	return nil
}

func responseChallengeCallback(w http.ResponseWriter, r *http.Request) (err error) {
	log.Debug("responseChallengeCallback")
	var response SubscriptionCallback
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		log.Fatal("Error decoding response body", err)
		return err
	}
	w.WriteHeader(http.StatusOK)
	if response.Challenge != "" {
		log.Debug("responseChallengeCallback: ", response)
		w.Write([]byte(response.Challenge))
		return
	} else {
		log.Debug("Sending to : ", response)
		for _, con := range c {
			con.WriteJSON(response)
		}
	}
	w.Write([]byte("OK"))
	return nil
}

func createSubscription(
	subcriptionType string,
	token string,
	channelId string,
	callbackUrl string,
	secret string) (subscription SubsciptionCreatedResponse, err error) {

	var jsonStr = []byte(fmt.Sprintf(`
        {
                "type": "%s",
                "version": "1",
                "condition": {
                        "broadcaster_user_id": "%s"
                },
                "transport": {
                        "method": "webhook",
                        "callback": "%s/callback",
                        "secret": "%s"
                }
        }`, subcriptionType, channelId, callbackUrl, secret))
	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/eventsub/subscriptions", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Error(err)
		return SubsciptionCreatedResponse{}, err
	}
	req.Header.Set("Client-ID", clientID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal("Creating subscription", err)
		return SubsciptionCreatedResponse{}, err
	}
	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return SubsciptionCreatedResponse{}, err
	}
	var response struct {
		Data []SubsciptionCreatedResponse `json:"data"`
	}

	log.Debug("Subscription created: ", string(bodyData))
	err = json.Unmarshal(bodyData, &response)
	if err != nil {
		log.Error(err)
		return SubsciptionCreatedResponse{}, err
	}

	return response.Data[0], nil
}

func removeAllSubs(token string, channelId string) (err error) {
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
	// log.Debugstring(responseData))
	return nil
}
