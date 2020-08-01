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
		help(s, m)
		deleteMessage()
		return
	}

	if len(m.Mentions) != 1 {
		return
	}
	if m.Mentions[0].ID != s.State.User.ID {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			L.Printf("Recovering from error '%s'.", r)

			msg := discordgo.MessageSend{
				Embed: &discordgo.MessageEmbed{Color: 0xff0000, Description: fmt.Sprintf("Something went wrong with your request `%s`.\n"+
					"Please try again laster or contact the server administrator.", replaceMentions(m.Content, s)),
					Title: "Error"},
			}
			_, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)
			if err != nil {
				L.Fatal(err)
			}
		}
	}()

	if callback := CommandList.getCallback(m); callback != nil {
		callback(s, m)
	} else {
		return
	}
	deleteMessage()
}

func replaceMentions(message string, s *discordgo.Session) string {
	message = strings.Replace(message, fmt.Sprintf("<@!%s>", s.State.User.ID), "@"+s.State.User.Username, -1)
	message = strings.Replace(message, fmt.Sprintf("<@%s>", s.State.User.ID), "@"+s.State.User.Username, -1)
	return message
}
