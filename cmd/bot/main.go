package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/IcelandicIcecream/go-discord-music-bot/internal/bot"
	"github.com/IcelandicIcecream/go-discord-music-bot/internal/config"
	"github.com/IcelandicIcecream/go-discord-music-bot/internal/services"
)

func main() {

	// Load config from viper
	config, err := config.LoadConfig()
	if err != nil {
		fmt.Println("error loading config,", err)
		return
	}

	// Initialize Youtube Service
	youtubeService, err := services.NewYouTubeService(config.YoutubeToken)
	if err != nil {
		fmt.Println("error creating Youtube service, ", err)
		return
	}

	// Create a new Discord session using the provided bot token.
	bot, err := bot.NewBot(config.DiscordToken, youtubeService)
	if err != nil {
		fmt.Println("error creating bot,", err)
		return
	}

	// Register ready as a callback for the ready events.
	bot.Session.AddHandler(bot.ReadyHandler)

	// Register messageCreate as a callback for the messageCreate events.
	bot.Session.AddHandler(bot.MessageCreateHandler)

	// Open a websocket connection to Discord and begin listening.
	err = bot.Session.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	bot.Session.Close()
}
