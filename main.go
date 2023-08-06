package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		fmt.Println("error loading config,", err)
		return
	}

	discord, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register ready as a callback for the ready events.
	discord.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	discord.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	// Called when the bot is ready.
	fmt.Println("Bot is ready!")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignoring bot messages.
	if m.Author.ID == s.State.User.ID {
		return
	}

	fmt.Println("Message from", m.Author.Username, "saying", m.Content)
}
