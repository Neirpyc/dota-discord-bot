package main

import (
	"bytes"
	"fmt"
	"github.com/Neirpyc/dota2api"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HeroList []dota2api.Hero

type ItemList []dota2api.Item

const (
	invalid = iota - 1
	publicMatchmaking
	pratice
	tournament
	tutorial
	coopWithAI
	teamMatch
	soloQueue
	rankedMatchmaking
	soloMid1v1
)

var (
	lobbyTypeConvert = map[int]string{
		invalid:           "invalid",
		publicMatchmaking: "Unranked",
		pratice:           "Practice",
		tournament:        "Tournament",
		tutorial:          "Tutorial",
		coopWithAI:        "Co-op With AI",
		teamMatch:         "Team Ranked",
		soloQueue:         "Unranked",
		rankedMatchmaking: "Ranked",
		soloMid1v1:        "Solo Mid 1v1",
	}
)

func (h HeroList) getName(id int) string {
	for i := 0; i < len(h); i++ {
		if id == h[i].ID {
			return h[i].Name
		}
	}
	return ""
}

func (i ItemList) getName(id int) string {
	for c := 0; c < len(i); c++ {
		if id == i[c].Id {
			return i[c].Name
		}
	}
	return ""
}

func getMatchReplacement(match dota2api.MatchSummary, steamId string) Replacement {
	r := make(Replacement)

	//getting match details
	details, err := D.GetMatchDetails(match.MatchID)
	if err != nil {
		L.Fatal(err)
	}

	ordonedPlayers := make([]dota2api.Player, 10)
	radiantCount := 0
	direCount := 5
	for _, v := range details.Result.Players {
		if v.PlayerSlot < 128 { //if radiant
			ordonedPlayers[radiantCount] = v
			radiantCount++
		} else {
			ordonedPlayers[direCount] = v
			direCount++
		}
	}

	for i := 0; i < 10; i++ {
		r[fmt.Sprintf("hero_name_%d", i)] =
			strings.Replace(HeroList(Heroes).getName(ordonedPlayers[i].HeroID), "npc_dota_hero_", "", 1)
		if fmt.Sprintf("%d", ordonedPlayers[i].AccountID) == steamId {
			r[fmt.Sprintf("class_hero_%d", i)] = "player"
		} else {
			r[fmt.Sprintf("class_hero_%d", i)] = ""
		}

		r[fmt.Sprintf("kills_player_%d", i)] = fmt.Sprintf("%d", ordonedPlayers[i].Kills)
		r[fmt.Sprintf("assists_player_%d", i)] = fmt.Sprintf("%d", ordonedPlayers[i].Assists)
		r[fmt.Sprintf("deaths_player_%d", i)] = fmt.Sprintf("%d", ordonedPlayers[i].Deaths)

		//gold
		r[fmt.Sprintf("gold_player_%d", i)] = func(gold int) string {
			if gold < 1000 {
				return fmt.Sprintf("%d", gold)
			}
			return fmt.Sprintf("%d.%dk", gold/1000, (gold%100)/100)
		}(ordonedPlayers[i].Gold)

		//items
		r[fmt.Sprintf("player_%d_item_0", i)] = strings.Replace(ItemList(Items).getName(ordonedPlayers[i].HeroID), "item_", "", 1)
	}
	if details.Result.RadiantWin {
		r["radiant_win"] = "true"
		r["dire_win"] = "false"
	} else {
		r["radiant_win"] = "false"
		r["dire_win"] = "true"
	}
	r["radiant_score"] = fmt.Sprintf("%d", details.Result.RadiantScore)
	r["dire_score"] = fmt.Sprintf("%d", details.Result.DireScore)

	//time label
	d, err := time.ParseDuration(strconv.Itoa(details.Result.Duration) + "s")
	if err != nil {
		L.Fatal(err)
	}
	if int64(d.Seconds())%60 > 10 {
		r["match_length"] = fmt.Sprintf("%d:%d", int64(d.Minutes()), int64(d.Seconds())%60)
	} else {
		r["match_length"] = fmt.Sprintf("%d:%d0", int64(d.Minutes()), int64(d.Seconds())%60)
	}

	//game date
	timeBegin := time.Unix(int64(match.StartTime), 0)
	r["game_date"] = fmt.Sprintf("%2.2d/%2.2d/%4.4d", timeBegin.Month(), timeBegin.Day(), timeBegin.Year())
	r["game_type"] = lobbyTypeConvert[match.LobbyType]

	//items

	return r
}

func getMatchImgSmall(match dota2api.MatchSummary, steamId string) []image.Image {
	r := getMatchReplacement(match, steamId)

	path, err := r.applyTemplate("assets/templates/small.html")
	if err != nil {
		L.Println(err)
		return nil
	}
	imgData := screenshotFile(path, "#render")
	if err = os.Remove("assets/" + path); err != nil {
		L.Println(err)
	}
	if img, err := png.Decode(bytes.NewReader(imgData[0])); err != nil {
		L.Fatal(err)
		return nil
	} else {
		return []image.Image{img}
	}
}

func getMatchImgMedium(match dota2api.MatchSummary, steamId string) []image.Image {
	r := getMatchReplacement(match, steamId)

	path, err := r.applyTemplate("assets/templates/medium.html")
	if err != nil {
		L.Println(err)
		return nil
	}
	imgsData := screenshotFile(path, "#render0", "#render1", "#render2")
	if err = os.Remove("assets/" + path); err != nil {
		L.Println(err)
	}
	imgs := make([]image.Image, 3)
	for i, imgData := range imgsData {
		if imgs[i], err = png.Decode(bytes.NewReader(imgData)); err != nil {
			L.Panic(err)
		}
	}
	return imgs
}

const urlHeroes = "http://cdn.dota2.com/apps/dota2/images/heroes/<name>_<size>.<ext>"

var sizes = []string{"lg", "vert", "sb", "full"}
var exts = []string{"png", "jpg", "png", "png"}

func createHeroesImagesList() {
	heroesList, err := D.GetHeroes()
	if err != nil {
		L.Fatal(err)
	}

	wg := &sync.WaitGroup{}

	wg.Add(len(heroesList))
	for _, v := range heroesList {
		go func(name string, wg *sync.WaitGroup) {
			customUrl := strings.Replace(urlHeroes, "<name>", name, 1)

			cli := http.Client{}

			for i, size := range sizes {
				path := "assets/heroes/" + size + "/" + name + "." + exts[i]
				if _, err := os.Stat(path); !os.IsNotExist(err) && !Config.ForceReload {
					continue
				}

				sizedUrl := strings.Replace(customUrl, "<size>", size, 1)
				sizedUrl = strings.Replace(sizedUrl, "<ext>", exts[i], 1)
				res, err := cli.Get(sizedUrl)
				if err != nil {
					L.Fatal(err)
				}

				data, err := ioutil.ReadAll(res.Body)
				if err != nil {
					L.Fatal(err)
				}

				err = ioutil.WriteFile(path, data, 0666)
				if err != nil {
					L.Fatal(err)
				}
			}
			wg.Done()
		}(strings.ReplaceAll(v.Name, "npc_dota_hero_", ""), wg)
	}
	wg.Wait()
}

const urlItems = "http://cdn.dota2.com/apps/dota2/images/items/<name>_lg.png"

func createItemsImagesList() {
	itemsList, err := D.GetItems()
	if err != nil {
		L.Fatal(err)
	}

	wg := &sync.WaitGroup{}

	wg.Add(len(itemsList))
	for _, v := range itemsList {
		go func(name string, wg *sync.WaitGroup) {
			defer wg.Done()
			customUrl := strings.Replace(urlItems, "<name>", name, 1)

			cli := http.Client{}

			path := "assets/items/lg/" + name + ".png"
			if _, err := os.Stat(path); !os.IsNotExist(err) && !Config.ForceReload {
				return
			}
			res, err := cli.Get(customUrl)
			if err != nil {
				L.Fatal(err)
			}

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				L.Fatal(err)
			}

			err = ioutil.WriteFile(path, data, 0666)
			if err != nil {
				L.Fatal(err)
			}
		}(strings.ReplaceAll(v.Name, "item_", ""), wg)
	}
	wg.Wait()
}
