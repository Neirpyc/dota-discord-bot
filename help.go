package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

const helpMessage = "This is  a work in progress Dota bot.\n" +
	"__**Usage:**__\n" +
	"	`!help` : Shows this message\n" +
	"The following commands must be used with a <bot@> in the message:\n" +
	"	`help @<bot_name>` : Show this message\n" +
	"	`last @<bot_name>` : Show the last match of the user\n" +
	"	`register <steam_username> @<bot_name>` : Register your steam account so you can use the service"

func Help(s *discordgo.Session, m *discordgo.MessageCreate) {
	_ = s.ChannelTyping(m.ChannelID)

	_, err := GetSteamID(m.Author.ID)
	if err != nil {
		L.Println(err)
		return
	}
	msg := strings.Replace(helpMessage, "<bot_name>", s.State.User.Username, -1)
	msg = strings.Replace(msg, "<bot@>", "<@"+s.State.User.Username+"#"+s.State.User.Discriminator+">", -1)
	_, err = s.ChannelMessageSend(m.ChannelID, msg)
	if err != nil {
		L.Printf("Sending message failed with error: %S", err.Error())
	}
}
