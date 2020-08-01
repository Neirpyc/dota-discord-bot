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

	steamId, err := GetSteamID(m.Author.ID)
	if err != nil {
		L.Panic(fmt.Sprintf("Couldn't find you steam ID in our database. Please register it by using "+
			"`register <steam_id> @%s`", s.State.User.Username))
		return
	}

	sizes := []string{"small", "medium", "full"}
	size := "medium"
wordFor:
	for _, word := range strings.Split(m.Message.Content, " ") {
		for _, possibleSize := range sizes {
			if word == possibleSize {
				size = word
				break wordFor
			}
		}
	}

	val, err := D.GetMatchHistory(map[string]interface{}{
		"account_id":        steamId,
		"matches_requested": 1,
		"min_players":       "10",
	})

	if err != nil {
		if val.Result.Status == 15 {
			L.Panic("We cannot retrieve match information unless you allow it in your Dota profile")
		} else {
			L.Panic(err)
		}
	}

	if len(val.Result.Matches) != 1 {
		_, err = s.ChannelMessageSend(m.ChannelID, "We couldn't find any match.")
		if err != nil {
			L.Fatal("Sending message failed with error: %S", err.Error())
		}
	}

	var imgs []image.Image

	switch size {
	case "small":
		imgs = getMatchImgSmall(val.Result.Matches[0], steamId)
	case "medium":
		imgs = getMatchImgMedium(val.Result.Matches[0], steamId)
	case "full":
		imgs = getMatchImgSmall(val.Result.Matches[0], steamId)
	}

	var wr bytes.Buffer
	for i, img := range imgs {
		err = jpeg.Encode(&wr, img, &jpeg.Options{Quality: 100})
		if err != nil {
			L.Fatal(err)
		}

		file := discordgo.File{Name: fmt.Sprintf("%d_%s_%d.jpg", val.Result.Matches[0].MatchID, size, i), ContentType: "image/jpeg", Reader: &wr}

		msg := discordgo.MessageSend{Files: []*discordgo.File{&file}}

		_, err = s.ChannelMessageSendComplex(m.ChannelID, &msg)
		if err != nil {
			L.Printf("Sending message failed with error: %s", err.Error())
		}
	}
}
