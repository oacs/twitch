package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	irc "github.com/thoj/go-ircevent"
	"log"
	"net/http"
	"net/url"
	"os"
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
	godotenv.Load(".env.local")
	secret_key := os.Getenv("TWITCH_SECRET")
	nick := "oacs69"
	con := irc.IRC(nick, "IRCTestSSL")
	con.VerboseCallbackHandler = true
	con.Debug = true
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
	con.AddCallback("366", func(e *irc.Event) {
	})

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
	fs := http.FileServer(http.Dir("./static"))
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/chat", chat)
	http.Handle("/", fs)
	log.Fatal(http.ListenAndServe(*addr, nil))
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	// IRC init
	con.Loop()
}
