package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/calvinbui/slack-csgo/rcon"

	"github.com/autom8ter/slasher"
	"github.com/nlopes/slack"
)

func executeRCON(csgoCmd string, remoteConsole rcon.RCON) slasher.HandlerFunc {
	return func(s *slasher.Slasher, client *slack.Client, command *slack.SlashCommand) (i interface{}, err error) {
		response := remoteConsole.Execute(fmt.Sprintf("%s %s", csgoCmd, command.Text))
		return &slack.Message{
			Msg: slack.Msg{
				Text: string(response),
			},
		}, nil
	}
}

func main() {
	hostPort := os.Getenv("RCON_HOST")
	password := os.Getenv("RCON_PASS")
	slackToken := os.Getenv("SLACK_TOKEN")

	remoteConsole, err := rcon.New(hostPort, password)
	if err != nil {
		log.Fatalf("Failed to create a connection to the CSGO Server: %s", err)
	}

	slash := slasher.NewSlasher(slackToken)

	slash.AddHandler("/map", executeRCON("changemap", remoteConsole))
	slash.AddHandler("/reset", executeRCON("restart", remoteConsole))

	mux := http.NewServeMux()
	mux.Handle("/slasher", slash.HandlerFunc())

	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("failed to start server: %s", err.Error())
		os.Exit(1)
	}
}
