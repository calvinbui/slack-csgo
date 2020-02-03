package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ikkerens/rcon"
	"github.com/nlopes/slack"
)

var (
	remoteConsole     *rcon.Conn
	rconHost          = os.Getenv("RCON_HOST")
	rconPass          = os.Getenv("RCON_PASS")
	slackSigningToken = os.Getenv("SLACK_SIGNING_SECRET")
	slackChannel      = os.Getenv("SLACK_CHANNEL")
	err               error
)

func main() {
	// create RCON client to CSGO server
	remoteConsole, err = rcon.New(rconHost, rconPass)
	if err != nil {
		log.Fatalf("Failed to create a connection to the CSGO Server: %s", err)
	}
	fmt.Println("RCON client created")

	http.HandleFunc("/slack", func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, slackSigningToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = verifier.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if s.ChannelName != slackChannel {
			slackMsg(w, fmt.Sprintf("This application can only run in #%s", slackChannel))
			return
		}

		switch s.Command {
		case "/map":
			rconResponse, err := rconSend(fmt.Sprintf("changelevel %s", s.Text))
			if err != nil {
				slackMsg(w, fmt.Sprintf("An error occured: %s", err.Error()))
			} else if strings.Contains(rconResponse, "CModelLoader::Map_IsValid") {
				slackMsg(w, fmt.Sprintf("%s is not a valid map name", s.Text))
			} else {
				slackMsg(w, fmt.Sprintf("The map has been changed to %s", s.Text))
			}
		case "/restart":
			_, err := rconSend("restart")
			if err != nil {
				slackMsg(w, fmt.Sprintf("An error occured: %s", err.Error()))
			} else {
				slackMsg(w, "The game will restart at the end of the round")
			}
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("Web server listening")

	http.ListenAndServe(":3000", nil)
}

func rconSend(s string) (string, error) {
	fmt.Printf("\nExecute RCON command: %s", s)
	rconResponse, err := remoteConsole.Send(s)
	if err != nil {
		fmt.Printf("\nAn error occured: %s", err.Error())
		return "", err
	}
	fmt.Printf("\nThe server responsed with %s", rconResponse)
	return rconResponse, nil
}

func slackMsg(w http.ResponseWriter, msg string) {
	params := &slack.Msg{Text: msg, ResponseType: "in_channel"}
	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
