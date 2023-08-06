package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) ReadyHandler(s *discordgo.Session, event *discordgo.Ready) {
	// Called when the bot is ready.
	fmt.Println("Bot is ready!")
}

func (b *Bot) MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Printf("Message: %+v\n", m.Message)
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "Ping!" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
}
