package bot

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

func (b *Bot) ReadyHandler(s *discordgo.Session, event *discordgo.Ready) {
	// Called when the bot is ready.
	fmt.Println("Bot is ready!")

	b.userVoiceChannelMap = make(map[string]string)
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

	case "!play":
		if len(args) > 0 {

			// Find the channel that the message came from.
			c, err := s.State.Channel(m.ChannelID)
			if err != nil {
				// Could not find channel.
				return
			}

			// Find the guild for that channel.
			g, err := s.State.Guild(c.GuildID)
			if err != nil {
				// Could not find guild.
				return
			}

			// Look for the message sender in that guild's current voice states.
			for _, vs := range g.VoiceStates {
				if vs.UserID == m.Author.ID {
					b.userVoiceChannelMapLock.Lock()
					b.userVoiceChannelMap[m.Author.ID] = vs.ChannelID
					voiceChannelID := b.userVoiceChannelMap[m.Author.ID]
					b.userVoiceChannelMapLock.Unlock()

					// Get youtube ID from search query
					query := strings.Join(args, " ")
					response, err := b.YoutubeService.SearchVideos(query, 1)
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, "Error searching videos.")
						fmt.Println("Error searching videos: ", err)
						return
					}

					// If no videos found, return no videos found message
					if len(response.Items) == 0 {
						s.ChannelMessageSend(m.ChannelID, "No videos found.")
						return
					}

					videoID := response.Items[0].Id.VideoId
					videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

					// Join the voice channel defeaned but not muted
					voiceConnection, err := s.ChannelVoiceJoin(
						m.GuildID,
						voiceChannelID,
						false,
						true,
					)
					if err != nil {
						fmt.Printf("Error joining the voice channel: %v", err)
						return
					}
					// If a song is currently playing or if the queue is not empty, add the song to the queue.
					if b.GetIsPlaying() || !b.IsQueueEmpty(m.GuildID) {
						// Add the song to the queue and update via message
						s.ChannelMessageSend(
							m.ChannelID,
							"Song added to queue! 🎶 - "+videoURL,
						)
						b.AddToQueue(m.GuildID, videoURL)
					} else {
						// Add the song to the queue and immediately play it.
						s.ChannelMessageSend(m.ChannelID, "Now playing! 🎶 - "+videoURL)
						b.AddToQueue(m.GuildID, videoURL)
						go b.PlayFromQueue(voiceConnection, m.GuildID)
					}
				}
			}

			if _, exists := b.userVoiceChannelMap[m.Author.ID]; !exists {
				s.ChannelMessageSend(m.ChannelID, "You are not in a voice channel.")
			} else {
			}

		} else {
			s.ChannelMessageSend(m.ChannelID, "Please provide a search query.")
		}

	case "!skip":
		if controlChan, ok := b.controlChan[m.GuildID]; ok {
			// Send a skip signal to the control channel
			controlChan <- "skip"
			s.ChannelMessageSend(m.ChannelID, "Skipped the current song.")
		} else {
			s.ChannelMessageSend(m.ChannelID, "No song is currently playing.")
		}

	case "!stop":
		if controlChan, ok := b.controlChan[m.GuildID]; ok {
			// Send a stop signal to the control channel
			controlChan <- "stop"
			s.ChannelMessageSend(m.ChannelID, "Stopped playback.")
		} else {
			s.ChannelMessageSend(m.ChannelID, "No song is currently playing.")
		}

	default:
		if m.Content == "Ping!" {
			s.ChannelMessageSend(m.ChannelID, "Pong!")
		}
	}
}

func (b *Bot) VoiceStateUpdateHandler(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	b.userVoiceChannelMapLock.Lock()
	defer b.userVoiceChannelMapLock.Unlock()
	b.userVoiceChannelMap[vsu.UserID] = vsu.ChannelID
}

// downloadAudio downloads an audio from a given URL and returns its path.
func (b *Bot) DownloadAudio(videoURL string) (string, error) {
	fmt.Println("downloading audio file...")

	if _, err := os.Stat("tmp/"); os.IsNotExist(err) {
		if err := os.Mkdir("tmp/", 0755); err != nil {
			return "", fmt.Errorf("error creating tmp directory: %v", err)
		}
	}

	// Generate a UUID for the temp filename
	fileUUID := uuid.New().String()
	audioFilePath := fmt.Sprintf("tmp/%s.mp3", fileUUID)

	cmd := exec.Command("youtube-dl", "-x", "--audio-format", "mp3", "-o", audioFilePath, videoURL)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("youtube-dl error:", stderr.String())
		return "", err
	}

	fmt.Println("Download Complete: ", audioFilePath)

	return audioFilePath, nil
}

func (b *Bot) PlayAudioFile(
	vc *discordgo.VoiceConnection,
	videoURL string,
	doneChan chan bool,
	interruptChan chan bool,
) error {
	audioFilePath, err := b.DownloadAudio(videoURL)
	if err != nil {
		return fmt.Errorf("error downloading audio file: %v", err)
	}

	// Cleanup after playing.
	defer func() {
		err = os.Remove(audioFilePath)
		if err != nil {
			fmt.Printf("Failed to remove audio file: %v\n", err)
		}
	}()

	return b.PlayAudioInVC(vc, audioFilePath, doneChan, interruptChan)
}

func (b *Bot) PlayAudioInVC(
	vc *discordgo.VoiceConnection,
	audioFilePath string,
	doneChan chan bool,
	interruptChan chan bool,
) error {
	b.wg.Add(1)
	defer b.wg.Done()
	fmt.Println("Playing audio file...")

	stopPlaying := make(chan bool) // Channel to stop the playback

	// Start playing the audio in a separate goroutine
	go func() {
		dgvoice.PlayAudioFile(vc, audioFilePath, stopPlaying)
		doneChan <- true
	}()

	select {
	case <-interruptChan:
		// Stop the playback by sending a signal to stopPlaying
		stopPlaying <- true
		b.SetIsPlaying(false) // Update IsPlaying flag
	case <-doneChan:
		// Playback finished, do nothing
	}

	return nil
}

func (b *Bot) AddToQueue(guildID string, videoURL string) {
	b.queueLock.Lock()
	defer b.queueLock.Unlock()

	b.queue[guildID] = append(b.queue[guildID], videoURL)
	fmt.Println("Added to queue: ", videoURL)
}

// GetNextInQueue gets the next song from the queue
func (b *Bot) GetNextInQueue(guildID string) (string, bool) {
	b.queueLock.Lock()
	defer b.queueLock.Unlock()

	if len(b.queue[guildID]) == 0 {
		return "", false
	}

	nextURL := b.queue[guildID][0]
	b.queue[guildID] = b.queue[guildID][1:]

	return nextURL, true
}

func (b *Bot) IsQueueEmpty(guildID string) bool {
	b.queueLock.Lock()
	defer b.queueLock.Unlock()

	return len(b.queue[guildID]) == 0
}

// func (b *Bot) PlayFromQueue(vc *discordgo.VoiceConnection, guildID string) {
// 	for {
// 		if b.IsQueueEmpty(guildID) {
// 			b.SetIsPlaying(false)
// 			break
// 		}
//
// 		nextSong, _ := b.GetNextInQueue(guildID)
//
// 		doneChan := make(chan bool)
//
// 		b.SetIsPlaying(true)
// 		go func() {
// 			defer func() {
// 				close(doneChan)
// 				b.SetIsPlaying(false)
// 			}()
//
// 			err := b.PlayAudioFile(vc, nextSong, doneChan)
// 			if err != nil {
// 				fmt.Printf("Error playing audio: %v", err)
// 				vc.Disconnect()
// 				return
// 			}
//
// 			<-doneChan
// 		}()
//
// 		// Wait for the current song to finish before starting the next one.
// 		<-doneChan
// 	}
//
// 	// Disconnect from the voice channel
// 	vc.Disconnect()
// }

func (b *Bot) PlayFromQueue(vc *discordgo.VoiceConnection, guildID string) {
	// Create a WaitGroup to wait for all playback goroutines to finish
	var wg sync.WaitGroup

	for {
		if b.IsQueueEmpty(guildID) {
			b.SetIsPlaying(false)
			break
		}

		interruptChan := make(chan bool)

		nextSong, _ := b.GetNextInQueue(guildID)

		doneChan := make(chan bool)
		controlChan := make(chan string)

		b.SetIsPlaying(true)

		b.controlChan[guildID] = controlChan // Store controlChan in the map

		// Increment the WaitGroup counter for the new playback goroutine
		wg.Add(1)

		go func() {
			defer func() {
				close(doneChan)
				b.SetIsPlaying(false)
				wg.Done() // Decrement the WaitGroup counter when the goroutine completes
			}()

			err := b.PlayAudioFile(vc, nextSong, doneChan, interruptChan)
			if err != nil {
				fmt.Printf("Error playing audio: %v", err)
				vc.Disconnect()
				return
			}

			<-doneChan
		}()

		select {
		case <-doneChan:
		case signal := <-controlChan:
			// Handle control signals
			if signal == "skip" {
				fmt.Println("Skip signal received. Skipping current song.")
				interruptChan <- true
				continue // Skip to the next iteration of the PlayFromQueue loop
			} else if signal == "stop" {
				fmt.Println("Stop signal received. Stopping playback.")
				interruptChan <- true
				b.ClearQueue(guildID)
				vc.Disconnect()
				return // Exit the PlayFromQueue loop
			}
		}
	}

	// Wait for all playback goroutines to finish before removing controlChan
	wg.Wait()

	// Remove controlChan from the map after all songs have finished
	delete(b.controlChan, guildID)

	// Disconnect from the voice channel
	vc.Disconnect()
}

func (b *Bot) SetIsPlaying(val bool) {
	if val {
		atomic.StoreInt32(&b.isPlaying, 1)
	} else {
		atomic.StoreInt32(&b.isPlaying, 0)
	}
}

func (b *Bot) ClearQueue(guildID string) {
	b.queueLock.Lock()
	defer b.queueLock.Unlock()

	b.queue[guildID] = []string{}
}

func (b *Bot) GetIsPlaying() bool {
	return atomic.LoadInt32(&b.isPlaying) == 1
}
