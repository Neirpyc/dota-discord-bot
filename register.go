package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

func register(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	_ = s.ChannelTyping(m.ChannelID)
	message := ""
	if len(args) != 2 {
		message = "Too many arguments.\n Try `help @<bot_name>` for more informations"
	}

	//todo use login instead of ID
	id, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		message = fmt.Sprintf("Couldn't parse steam ID `%s`.", args[1])
	} else {
		err = SetSteamID(m.Author.ID, fmt.Sprintf("%d", id))
		if err != nil {
			L.Fatal(err)
		}

		message = fmt.Sprintf("Your steam ID is `%d`. You may now use our service.", id)
	}

	message = strings.Replace(message, "<bot_name>", s.State.User.Username, -1)

	_, err = s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		L.Printf("Sending message failed with error: %S", err.Error())
	}
}
