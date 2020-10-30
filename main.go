package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// settings
var (
	name   = "CatBot"
	color  = 0x00adef
	prefix = "!"
)

// Cat struct for cat data
type Cat struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Origin       string `json:"origin"`
	Temperament  string `json:"temperament"`
	WikipediaURL string `json:"wikipedia_url"`
}

// init func called when running bot
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error load .env file")
	}
}

func main() {

	// Create a new discord session using the provide bot token
	discord, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
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
	if strings.HasPrefix(m.Content, prefix+"help") {
		help := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{Name: name + " Commands"},
			Color:  color,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   prefix + "help",
					Value:  "Display a list of commands.",
					Inline: true,
				},
				{
					Name:   prefix + "hello",
					Value:  "send message Hello, World",
					Inline: true,
				},
				{
					Name:   prefix + "cat",
					Value:  "Display a random cat image.",
					Inline: true,
				},
			},
		}
		s.ChannelMessageSendEmbed(m.ChannelID, help)
	}

	// Return reply when the message is "!hello"
	if strings.HasPrefix(m.Content, prefix+"hello") {
		s.ChannelMessageSend(m.ChannelID, "Hello,World!")
	}

	// Get image from Cat Api
	if strings.HasPrefix(m.Content, prefix+"cat") {
		req, err := http.NewRequest("GET", "http://thecatapi.com/api/images/get?", nil)
		req.Header.Set("x-api-key", os.Getenv("APIKEY"))
		if err != nil {
			log.Fatal("Error:", err.Error())
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Error:", err.Error())
		}

		defer resp.Body.Close()

		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Unable to get cat!")
		} else {
			s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					Name: "Cat Image",
				},
				Color: color,
				Image: &discordgo.MessageEmbedImage{
					URL: resp.Request.URL.String(),
				},
			})
		}
	}

	// Search cat from Cat Api
	if strings.HasPrefix(m.Content, prefix+"search ") {
		req, err := http.NewRequest("GET", "https://api.thecatapi.com/v1/breeds/search?q="+m.Content[len(prefix)+7:], nil)
		req.Header.Set("x-api-key", os.Getenv("APIKEY"))
		if err != nil {
			log.Fatal("Error:", err.Error())
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Error:", err.Error())
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error:", err.Error())
		}

		var cat []Cat
		err = json.Unmarshal(body, &cat)
		if err != nil {
			log.Fatal("Error:", err.Error())
		}

		if len(cat) <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Cat not found!")
		} else {
			for _, each := range cat {
				s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
					Author: &discordgo.MessageEmbedAuthor{
						Name: each.Name,
					},
					Color:       color,
					Description: each.Description,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Origin",
							Value: each.Origin,
						},
						{
							Name:  "Temperament",
							Value: each.Temperament,
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: each.WikipediaURL,
					},
				})
			}
		}
	}

}
