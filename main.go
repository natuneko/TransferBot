package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var token string

type Config struct {
	Token       string `json:"token"`
	TranferFrom string `json:"tranferFrom"`
	TransferTo  string `json:"transferTo"`
	MainRoom    string `json:"mainRoom"`
}

var config Config

func init() {
	configfile, _ := ioutil.ReadFile("config.json")
	json.Unmarshal(configfile, &config)
	token = config.Token
}

type Webhook struct {
	AvatarURL string `json:"avatar_url"`
	Username  string `json:"username"`
	Content   string `json:"content"`
}

func SendWebhook(url string, content string, username string, avatarURL string) {
	client := &http.Client{}

	var webhook Webhook

	webhook.AvatarURL = avatarURL
	webhook.Username = username
	webhook.Content = content

	data, _ := json.Marshal(&webhook)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Add("content-type", "application/json")
	client.Do(req)
}

func main() {
	client, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	client.AddHandler(onReady)
	client.AddHandler(onMessage)

	err = client.Open()
	if err != nil {
		panic(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	<-sc
	client.Close()
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	fmt.Println("Bot Start")
	fmt.Println("UserID:", r.User.ID)
	fmt.Println("UserName:", r.User.Username)
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID == config.TranferFrom {
		for _, mention := range m.Mentions {
			m.Content = strings.ReplaceAll(m.Content, fmt.Sprintf("<@%s>", mention.ID), fmt.Sprintf("`@%s`", mention.Username))
		}

		for _, role := range m.MentionRoles {
			role, _ := s.State.Role(m.GuildID, role)
			m.Content = strings.ReplaceAll(m.Content, fmt.Sprintf("<@&%s>", role.ID), fmt.Sprintf("`@%s`", role.Name))
		}

		m.Content = strings.ReplaceAll(m.Content, "@everyone", "`@everyone`")
		m.Content = strings.ReplaceAll(m.Content, "@here", "`@here`")

		for _, attachment := range m.Attachments {
			m.Content += "\n" + attachment.URL
		}

		if m.ChannelID != config.MainRoom {
			channel, _ := s.Channel(m.ChannelID)
			m.Content += " - " + channel.Name
		}

		SendWebhook(config.TransferTo, m.Content, m.Author.Username, m.Author.AvatarURL(""))
	}
}
