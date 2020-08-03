package main

import (
	"bytes"
	"fmt"
	"github.com/Neirpyc/dota2api"
	"image"
	"image/png"
	"os"
	"strconv"
	"sync"
	"time"
)

func getMatchReplacement(match dota2api.MatchSummary, steamId string) Replacement {
	r := make(Replacement)

	//getting match details
	details, err := D.GetMatchDetails(match.MatchId)
	if err != nil {
		L.Fatal(err)
	}

	orderedPlayers := make([]dota2api.PlayerJSON, 10)
	radiantCount := 0
	direCount := 5
	for _, v := range details.Result.Players {
		if v.PlayerSlot < 128 { //if radiant
			orderedPlayers[radiantCount] = v
			radiantCount++
		} else {
			orderedPlayers[direCount] = v
			direCount++
		}
	}

	for i := 0; i < 10; i++ {
		h, found := Heroes.GetById(orderedPlayers[i].HeroID)
		if !found {
			L.Panic("Hero not found")
		}
		r[fmt.Sprintf("hero_name_%d", i)] = h.Name.GetName()
		if fmt.Sprintf("%d", orderedPlayers[i].AccountID) == steamId {
			r[fmt.Sprintf("class_hero_%d", i)] = "player"
		} else {
			r[fmt.Sprintf("class_hero_%d", i)] = ""
		}

		r[fmt.Sprintf("kills_player_%d", i)] = fmt.Sprintf("%d", orderedPlayers[i].Kills)
		r[fmt.Sprintf("assists_player_%d", i)] = fmt.Sprintf("%d", orderedPlayers[i].Assists)
		r[fmt.Sprintf("deaths_player_%d", i)] = fmt.Sprintf("%d", orderedPlayers[i].Deaths)

		//gold
		r[fmt.Sprintf("gold_player_%d", i)] = func(gold int) string {
			if gold < 1000 {
				return fmt.Sprintf("%d", gold)
			}
			return fmt.Sprintf("%d.%dk", gold/1000, (gold%100)/100)
		}(orderedPlayers[i].Gold)

		//items
		i, found := Items.GetById(orderedPlayers[i].Item0)
		if !found {
			L.Panic("Item not found")
		}
		r[fmt.Sprintf("player_%d_item_0", i)] = i.Name.GetName()
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
	r["game_date"] = fmt.Sprintf("%2.2d/%2.2d/%4.4d", match.StartTime.Month(), match.StartTime.Day(), match.StartTime.Year())
	r["game_type"] = match.LobbyType.GetName()

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

func createHeroesImagesList() {
	Heroes.GoForEach(func(hero dota2api.Hero, wg *sync.WaitGroup) {
		defer wg.Done()
		sizes := []string{"lg", "sb", "full", "vert"}
		for i, size := range sizes {

			path := "assets/heroes/" + size + "/" + hero.Name.GetName() + ".png"
			if _, err := os.Stat(path); !os.IsNotExist(err) && !Config.ForceReload {
				continue
			}

			img, err0 := D.GetHeroImage(hero, i)
			if err0 != nil {
				fmt.Println(hero)
				L.Fatal(err0)
			}

			f, err := os.Create(path)
			if err != nil {
				L.Fatal(err)
			}

			if err := png.Encode(f, img); err != nil {
				L.Fatal(err)
			}
		}
	})
}

func createItemsImagesList() {
	Items.GoForEach(func(item dota2api.Item, wg *sync.WaitGroup) {
		defer wg.Done()

		path := "assets/items/lg/" + item.Name.GetName() + ".png"
		if _, err := os.Stat(path); !os.IsNotExist(err) && !Config.ForceReload {
			return
		}

		img, err := D.GetItemImage(item)
		if err != nil {
			L.Fatal(err)
		}

		f, err := os.Create(path)
		if err != nil {
			L.Fatal(err)
		}

		err = png.Encode(f, img)
		if err != nil {
			L.Fatal(err)
		}
	})
}
