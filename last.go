package main

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"image"
	"image/jpeg"
	"strings"
)

func Last(s *discordgo.Session, m *discordgo.MessageCreate) {
	_ = s.ChannelTyping(m.ChannelID)
	message := "last :("

	steamId, err := GetSteamID(m.Author.ID)
	if err != nil {
		message = "Couldn't find you steam ID in our database. Please register it by using `register <steam_id> @<bot_name>`"
		message = strings.Replace(message, "<bot_name>", s.State.User.Username, -1)

		msg := discordgo.MessageSend{
			Embed: &discordgo.MessageEmbed{Color: 0xff0000, Description: fmt.Sprintf("Something went wrong with your request `%s`.\n"+
				"Please try again laster or contact the server administrator.\n Error is:\"%s\"", m.Content, message), Title: "Error"},
		}
		_, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)
		if err != nil {
			L.Fatal(err)
		}

		return
	}

	val, err := D.GetMatchHistory(map[string]interface{}{
		"account_id":        steamId,
		"matches_requested": 1,
		"min_players":       "10",
	})
	if err != nil {
		L.Println(err)
	}

	size := "small"

	var img image.Image

	if len(val.Result.Matches) != 1 {
		_, err = s.ChannelMessageSend(m.ChannelID, "We couldn't find any match.")
		if err != nil {
			L.Fatal("Sending message failed with error: %S", err.Error())
		}
	}

	img = getMatchImgSmall(val.Result.Matches[0], steamId)

	wr := bytes.Buffer{}

	err = jpeg.Encode(&wr, img, &jpeg.Options{Quality: 100})
	if err != nil {
		L.Fatal(err)
	}

	file := discordgo.File{Name: fmt.Sprintf("%d_%s.jpg", val.Result.Matches[0].MatchID, size), ContentType: "image/jpeg", Reader: &wr}

	msg := discordgo.MessageSend{Files: []*discordgo.File{&file}}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, &msg)
	if err != nil {
		L.Printf("Sending message failed with error: %S", err.Error())
	}
}
