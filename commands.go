package main

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"image"
	"image/jpeg"
	"strconv"
	"strings"
)

var (
	Help = Command{
		MustContain: []string{"help"},
		Priority:    1,
		Callback:    help,
	}
	Register = Command{
		MustContain: []string{"register"},
		Priority:    0,
		Callback:    register,
	}
	Last = Command{
		MustContain: []string{"last"},
		Priority:    0,
		Callback:    last,
	}
	CommandList = Commands{Help, Register, Last}
)

const (
	helpMessage = "This is  a work in progress Dota bot.\n" +
		"__**Usage:**__\n" +
		"	`!help` : Shows this message\n" +
		"The following commands must be used with a <bot@> in the message:\n" +
		"	`help @<bot_name>` : Show this message\n" +
		"	`last @<bot_name>` : Show the last match of the user\n" +
		"	`register <steam_username> @<bot_name>` : Register your steam account so you can use the service"
)

type Command struct {
	MustContain []string
	Priority    int
	Callback    func(*discordgo.Session, *discordgo.MessageCreate)
}

type Commands []Command

func (cs Commands) getCallback(m *discordgo.MessageCreate) func(*discordgo.Session, *discordgo.MessageCreate) {
	bestMatch := Command{
		Priority: -1,
	}
	ambiguous := false
commandLoop:
	for _, c := range cs {
		if c.Priority < bestMatch.Priority {
			continue
		}
	seekLoop:
		for _, seek := range c.MustContain {
			for _, word := range strings.Split(m.Content, " ") {
				if seek == word {
					continue seekLoop
				}
			}
			continue commandLoop
		}
		if c.Priority == bestMatch.Priority {
			ambiguous = true
		} else {
			ambiguous = false
		}
		bestMatch = c
	}
	if ambiguous {
		return nil
	}
	return bestMatch.Callback
}

func help(s *discordgo.Session, m *discordgo.MessageCreate) {
	_ = s.ChannelTyping(m.ChannelID)

	_, err := GetSteamID(m.Author.ID)
	if err != nil {
		L.Println(err)
		return
	}
	msg := strings.Replace(helpMessage, "<bot_name>", s.State.User.Username, -1)
	msg = strings.Replace(msg, "<bot@>", "`@"+s.State.User.Username+"`", -1)
	_, err = s.ChannelMessageSend(m.ChannelID, msg)
	if err != nil {
		L.Printf("Sending message failed with error: %S", err.Error())
	}
}

func register(s *discordgo.Session, m *discordgo.MessageCreate) {
	_ = s.ChannelTyping(m.ChannelID)
	message := ""

	//todo use login instead of ID
	var err error
	var id int64
	for _, word := range strings.Split(m.Content, " ") {
		if id, err = strconv.ParseInt(word, 10, 64); err != nil {
			continue
		} else {
			goto noError
		}
	}
	L.Panic("Couldn't parse your steam ID.")
noError:
	err = SetSteamID(m.Author.ID, fmt.Sprintf("%d", id))
	if err != nil {
		L.Fatal(err)
	}

	message = fmt.Sprintf("Your steam ID is `%d`. You may now use our service.", id)

	_, err = s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		L.Printf("Sending message failed with error: %s", err.Error())
	}
}

func last(s *discordgo.Session, m *discordgo.MessageCreate) {
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
