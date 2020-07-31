package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	deleteMessage := func() {
		err := s.ChannelMessageDelete(m.ChannelID, m.Message.ID)
		if err != nil {
			L.Printf("Failed to delete the original message with error: %S", err.Error())
			return
		}
	}

	if m.Content == "!help" {
		deleteMessage()
		Help(s, m)
		return
	}

	noMention := strings.Join(strings.Split(m.Content, "<@!"+s.State.User.ID+"> "), "")
	noMention = strings.Join(strings.Split(noMention, "<@!"+s.State.User.ID+">"), "")
	if noMention == m.Content {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			L.Printf("Recovering from error '%s'.", r)

			msg := discordgo.MessageSend{
				Embed: &discordgo.MessageEmbed{Color: 0xff0000, Description: fmt.Sprintf("Something went wrong with your request `%s`.\n"+
					"Please try again laster or contact the server administrator.", m.Content), Title: "Error"},
			}
			_, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)
			if err != nil {
				L.Fatal(err)
			}
		}
	}()

	args := strings.Split(noMention, " ")
	switch strings.Split(noMention, " ")[0] {
	case "help":
		Help(s, m)
	case "last":
		Last(s, m)
	case "register":
		register(s, m, args)
	default:
		return
	}
	deleteMessage()
}
