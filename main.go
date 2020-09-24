package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	Name   = "CatBot"
	Color  = 0x00adef
	Prefix = "!"
	// Token api
	Token  string
	CatAPI string
)

// init func called when running bot
func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	// Create a new discord session using the provide bot toen
	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("unable to create discord client", err)
	}

	// Register the read func as callback for message events
	discord.AddHandler(messageCreate)

	// Open websocket connection to discord and begin listening
	err = discord.Open()
	if err != nil {
		log.Fatal("unable to open connection to discord", err)
	}

	// Bot is now running
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Close down the discord session
	discord.Close()
}

// messageCreate func will be called to AddHandler everytime a new message created on any channel
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	client := &http.Client{}

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Display all list of commands
	if strings.HasPrefix(m.Content, Prefix+"help") {
		help := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{Name: Name + " Commands"},
			Color:  Color,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   Prefix + "help",
					Value:  "Display a list of commands.",
					Inline: true,
				},
				{
					Name:   Prefix + "hello",
					Value:  "send message Hello, World",
					Inline: true,
				},
				{
					Name:   Prefix + "cat",
					Value:  "Display a random cat picture.",
					Inline: true,
				},
			},
		}
		s.ChannelMessageSendEmbed(m.ChannelID, help)
	}

	// Return reply when the message is "!hello"
	if strings.HasPrefix(m.Content, Prefix+"hello") {
		s.ChannelMessageSend(m.ChannelID, "Hello,World!")
	}

	// Get picture from Cat Api
	if strings.HasPrefix(m.Content, Prefix+"cat") {
		req, err := http.NewRequest("GET", "http://thecatapi.com/api/images/get", nil)
		req.Header.Set("x-api-key", CatAPI)

		resp, err := client.Do(req)
		if err != nil {
			defer resp.Body.Close()
		}

		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Unable to get cat!")
		} else {
			s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					Name: "Cat Picture",
				},
				Color: Color,
				Image: &discordgo.MessageEmbedImage{
					URL: resp.Request.URL.String(),
				},
			})
		}
	}

}
