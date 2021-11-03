package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"github.com/gorilla/websocket"
	irc "github.com/thoj/go-ircevent"
)

// Websocket connection address
var addr = flag.String("addr", "localhost:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

// Websocket connections array
var c = []*websocket.Conn{}

func chat(w http.ResponseWriter, r *http.Request) {
	con, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	c = append(c, con)
	last := c[len(c)-1]
	defer last.Close()
	for {
		mt, message, err := last.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = last.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

// oauth:rujzw7slvcyc8b8q2kwcirxdfx0sa7
const channel = "#oacs69"
const serverssl = "irc.chat.twitch.tv:6697"

func main() {
	nick := "oacs69"
	con := irc.IRC(nick, "IRCTestSSL")
	con.VerboseCallbackHandler = false
	con.Debug = true
	con.UseTLS = true
	con.Password = "oauth:rujzw7slvcyc8b8q2kwcirxdfx0sa7"
	con.TLSConfig = &tls.Config{InsecureSkipVerify: true}

  // On connect join the channel
	con.AddCallback("001", func(e *irc.Event) { con.Join(channel) })

  // Return the message to the websocket
	con.AddCallback("PRIVMSG", func(e *irc.Event) {
		for _, c := range c {
			data := e.Nick + ": " + e.Message()
			c.WriteMessage(websocket.TextMessage, []byte(data))
		}
	})

	err := con.Connect(serverssl)

	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}

  // Websocket init
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/chat", chat)
	log.Fatal(http.ListenAndServe(*addr, nil))
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

  // IRC init
	con.Loop()
}
