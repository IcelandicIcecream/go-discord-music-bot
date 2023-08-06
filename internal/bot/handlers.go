package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) ReadyHandler(s *discordgo.Session, event *discordgo.Ready) {
	// Called when the bot is ready.
	fmt.Println("Bot is ready!")
}

func (b *Bot) MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Split the message content into command and arguments
	content := strings.Split(m.Content, " ")
	command := content[0]
	args := content[1:]

	switch command {
	case "!search":
		// If command is "!search", we'll use YouTube Service to search for videos.
		if len(args) > 0 {
			query := strings.Join(args, " ")
			response, err := b.YoutubeService.SearchVideos(query, 3)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error searching videos.")
				return
			}

			// Collect video titles
			titles := []string{}
			for _, item := range response.Items {
				titles = append(titles, item.Snippet.Title)
			}

			// Send the titles as a message
			s.ChannelMessageSend(m.ChannelID, strings.Join(titles, "\n"))
		} else {
			s.ChannelMessageSend(m.ChannelID, "Please provide a search query.")
		}
	default:
		if m.Content == "Ping!" {
			s.ChannelMessageSend(m.ChannelID, "Pong!")
		}
	}
}
