package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	irc "github.com/thoj/go-ircevent"
)

type MessageData struct {
	Message     string            `json:"message"`
	DisplayName string            `json:"displayName"`
	Tags        map[string]string `json:"tags"`
	Type        string            `json:"type"`
}

const (
	channel   = "#oacs69"
	serverssl = "irc.chat.twitch.tv:6697"
	nick      = "oacs69"
)

var (
	addr     = flag.String("addr", "localhost:7001", "http service address")
	upgrader = websocket.Upgrader{} // use default options
	c        = []*websocket.Conn{}
)

// The same json tags will be used to encode data into JSON
func chat(w http.ResponseWriter, r *http.Request) (err error) {
	log.Debug("Starting chat")
	con, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error upgrading connection")
		return err
	}

	// add connection to list
	c = append(c, con)
	return err
}

// oauth:rujzw7slvcyc8b8q2kwcirxdfx0sa7

// TODO betterjthis
func twitchIRCConnect() {
	secret_key := os.Getenv("TWITCH_CHAT_OAUTH")
	con := irc.IRC(nick, "IRCTestSSL")
	con.VerboseCallbackHandler = false
	con.Debug = false
	con.UseTLS = true
	con.Password = "oauth:" + secret_key
	con.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// On connect join the channel
	con.AddCallback("001", func(e *irc.Event) {
		con.SendRaw("CAP REQ :twitch.tv/membership\n")
		con.SendRaw("CAP REQ :twitch.tv/commands\n")
		con.SendRaw("CAP REQ :twitch.tv/tags\n")
		con.Join(channel)
	})

	// Return the message to the websocket
	con.AddCallback("PRIVMSG", func(e *irc.Event) {
		// make json with the tags on event
		messageData := &MessageData{
			Message:     e.Message(),
			DisplayName: e.Tags["display-name"],
			Tags:        e.Tags,
			Type:        "message",
		}
		for _, con := range c {
			con.WriteJSON(messageData)
		}
	})

	err := con.Connect(serverssl)

	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}

	// IRC init
	go con.Loop()
}
