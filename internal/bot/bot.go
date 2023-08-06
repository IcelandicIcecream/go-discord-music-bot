package bot

import (
	"github.com/IcelandicIcecream/go-discord-music-bot/internal/services"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session        *discordgo.Session
	YoutubeService *services.YouTubeService
}

func NewBot(token string, youtubeService *services.YouTubeService) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		Session:        session,
		YoutubeService: youtubeService,
	}

	return bot, nil
}
