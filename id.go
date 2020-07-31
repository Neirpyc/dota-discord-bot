package main

import (
	"errors"
	"fmt"
	"strconv"
)

func GetSteamID(discordID string) (string, error) {
	_, err := DB.Exec("USE dota_bot_discord")
	if err != nil {
		return "", err
	}
	res, err := DB.Query("SELECT steam_id FROM steam_id WHERE discord_id=" + discordID)
	if err != nil {
		return "", err
	}

	var value string
	for res.Next() {
		err = res.Scan(&value)
		if err != nil {
			return "", err
		}
		if value != "" {
			return value, nil
		}
	}

	return "", errors.New("Couldn't find the steam ID in the database.")
}

func SetSteamID(discordID, steamId string) error {
	steamIdUInt64, err := strconv.ParseInt(steamId, 10, 64)
	if err != nil {
		L.Fatal(err)
	}
	B32Max := int64(^uint32(0))
	if steamIdUInt64 > B32Max {
		steamId = fmt.Sprintf("%d", B32Max&steamIdUInt64)
	}

	_, err = DB.Exec("USE dota_bot_discord;")
	if err != nil {
		return err
	}

	_, err = DB.Exec(fmt.Sprintf("INSERT INTO dota_bot_discord.steam_id (discord_id, steam_id) VALUE (%s, %s);", discordID, steamId))
	if err != nil {
		_, err = DB.Exec(fmt.Sprintf("REPLACE INTO dota_bot_discord.steam_id (discord_id, steam_id) VALUE (%s, %s);", discordID, steamId))
		if err != nil {
			return err
		}
	}
	return nil
}
