package bot

import (
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/IcelandicIcecream/go-discord-music-bot/internal/services"
)

type Bot struct {
	Session                 *discordgo.Session
	YoutubeService          *services.YouTubeService
	userVoiceChannelMap     map[string]string
	userVoiceChannelMapLock sync.Mutex
	isPlaying               int32
	queue                   map[string][]string
	queueLock               sync.Mutex
	wg                      sync.WaitGroup
	controlChan             map[string]chan string
}

func NewBot(token string, youtubeService *services.YouTubeService) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		Session:             session,
		YoutubeService:      youtubeService,
		userVoiceChannelMap: make(map[string]string), // Initializing the map
		queue:               make(map[string][]string),
		controlChan:         make(map[string]chan string),
	}

	return bot, nil
}
