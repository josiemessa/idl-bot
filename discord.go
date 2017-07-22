package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"log"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	session          *discordgo.Session
	channelID        string
	bufferedMessages chan string
	lastMessage      time.Time
}

func (d *DiscordBot) Connect(token string, guild string) error {
	log.Println("Logging into discord")
	d.bufferedMessages = make(chan string)

	// Create a new Discord session using the provided bot token.
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("Error creating Discord session: ", err)
	}

	d.session = s
	err = d.setChannelID(guild)
	if err != nil {
		return err
	}

	// Register ready as a callback for the ready events.
	d.session.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	d.session.AddHandler(messageCreate)

	return nil

}

func (d *DiscordBot) setChannelID(guild string) error {
	guilds, err := d.session.UserGuilds(-1, "", "")
	if err != nil {
		return fmt.Errorf("Error finding guilds: %s", err)
	}
	for _, g := range guilds {
		if g.Name == guild {
			channels, err := d.session.GuildChannels(g.ID)
			if err != nil {
				return fmt.Errorf("Error finding channels: %s", err)
			}
			for _, c := range channels {
				if c.Name == "general" && c.Type == "text" {
					d.channelID = c.ID
				}
			}
		}
	}
	if d.channelID == "" {
		return fmt.Errorf("Could not find general in guild %s", guild)
	}
	return nil
}

func (d *DiscordBot) Run() {
	sc := make(chan os.Signal, 1)

	// Open the websocket and begin listening.
	err := d.session.Open()
	if err != nil {
		log.Fatalln("Error opening Discord session: ", err)
	}
	// Wait here until CTRL-C or other term signal is received.
	log.Println("Discord is now running.  Press CTRL-C to exit.")
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case <-sc:
			// Cleanly close down the Discord session.
			d.session.Close()
			return
		case msg := <-d.bufferedMessages:
			go time.AfterFunc(RATE_LIMIT, func() {
				// never notify as we'll have already handled this before buffering the message
				d.SendMessage(msg, false)
			})
		}
	}

}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateStatus(0, "justbotthings")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	kotlresponses := map[string]struct{}{
		"Kundalini":                                                                          struct{}{},
		"Stand aside or be trampled!":                                                        struct{}{},
		"A quest? A quest! A-questing I shall go!":                                           struct{}{},
		"No epic fail today, my friends, tis brilliantly clear that this one is in the bag.": struct{}{},
		"From Light's Keep, Light's Keeper rides!":                                           struct{}{},
		"Play Ezalor.":                                                                       struct{}{},
		"Play I ride with the light.":                                                        struct{}{},
		"Play The Light guides me on my quest.":                                              struct{}{},
		"Somebody once told me the dark was gonna roll me":                                   struct{}{},
	}

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is to "@botl"
	atbotl := fmt.Sprintf("<@%s>", s.State.User.ID)
	if strings.Contains(m.Content, atbotl) {
		var message string
		switch {
		case strings.Contains(m.Content, "which team am I in"):
			message = findTeams(m.Author.Username)
		case strings.Contains(m.Content, "when is my next match"):
			message = findFixtures(m.Author.Username)
		case strings.Contains(m.Content, "which team is"):
			playerID := findPlayerInMessage(m.Content)
			user, err := s.User(playerID)
			if err != nil {
				message = "I'm sorry I don't know. I have failed in my quest."
			} else {
				message = findTeams(user.Username)
			}
		default:
			for response := range kotlresponses {
				message = response
				break
			}
		}
		message = fmt.Sprintf("%s %s", m.Author.Mention(), message)
		s.ChannelMessageSend(m.ChannelID, message)
	} else {
		log.Println(m.Content)
	}
}

func findPlayerInMessage(msg string) string {
	parts := strings.Split(msg, "<@")
	if len(parts) < 2 {
		return ""
	}
	return strings.Split(parts[1], ">")[0]
}

func findTeams(user string) string {
	for _, t := range teams {
		for _, p := range t.Players {
			if p.DiscordName == user {
				return t.Name
			}
		}
	}
	return "I'm sorry I don't know. I have failed in my quest."
}

func findFixtures(user string) string {

	teamName := findTeams(user)
	now := time.Now()
	for _, f := range fixtures {
		if strings.Contains(f.Teams, teamName) {
			t, _ := time.Parse("02/01/2006 3pm", fmt.Sprintf("%s 8pm", f.Date))
			if now.Before(t) {
				return fmt.Sprintf("you are next playing %s at 8pm %s %s", f.Teams, t.Weekday(), f.Date)
			}
		}
	}
	return "I'm sorry I don't know. I have failed in my quest."
}

// Sends a message to the #general channel in the IDL server
// If notify is true, then this will notify the whole channel
// Note this will be called in a lot of go routines so do not
// manipulate any structs without using locks
func (d *DiscordBot) SendMessage(message string, notify bool) {
	// Prefix message with "#general"
	if notify {
		message = fmt.Sprintf("<#%s>%s", d.channelID, message)
	}

	// Calculte the time since the last message was sent so
	// we can rate limit ourselves
	sinceLastMessage := time.Since(d.lastMessage).Seconds()

	// There's a chance that we will call SendMessage() from
	// generating the fixture alerts before we're logged into
	// discord, so hold onto the messages
	if d.session != nil || sinceLastMessage > RATE_LIMIT.Seconds() {
		_, err := d.session.ChannelMessageSend(d.channelID, message)
		if err != nil {
			log.Printf("Error sending message <<%s>>: %s", message, err)
		}
		d.lastMessage = time.Now()
	} else {
		d.bufferedMessages <- message
	}
}
