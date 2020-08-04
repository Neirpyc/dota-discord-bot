package main

import (
	"bytes"
	"fmt"
	"github.com/Neirpyc/dota2api"
	"image"
	"image/png"
	"os"
	"sync"
)

func getMatchReplacement(match dota2api.MatchSummary, steamId string) Replacement {
	r := make(Replacement)

	//getting match details
	details, err := D.GetMatchDetails(dota2api.MatchId(match.MatchId))
	if err != nil {
		L.Fatal(err)
	}

	var i int
	details.ForEachPlayer(func(p dota2api.PlayerDetails) {
		r[fmt.Sprintf("hero_name_%d", i)] = p.Hero.Name.GetName()
		if fmt.Sprintf("%d", p.AccountId) == steamId {
			r[fmt.Sprintf("class_hero_%d", i)] = "player"
		} else {
			r[fmt.Sprintf("class_hero_%d", i)] = ""
		}

		r[fmt.Sprintf("kills_player_%d", i)] = fmt.Sprintf("%d", p.KDA.Kills)
		r[fmt.Sprintf("assists_player_%d", i)] = fmt.Sprintf("%d", p.KDA.Assists)
		r[fmt.Sprintf("deaths_player_%d", i)] = fmt.Sprintf("%d", p.KDA.Deaths)

		//gold
		r[fmt.Sprintf("gold_player_%d", i)] = p.Stats.Gold.NetWorth().ToString()

		//items
		r[fmt.Sprintf("player_%d_item_0", i)] = p.Items.Item0.Name.GetName()
		r[fmt.Sprintf("player_%d_item_1", i)] = p.Items.Item1.Name.GetName()
		r[fmt.Sprintf("player_%d_item_2", i)] = p.Items.Item2.Name.GetName()
		r[fmt.Sprintf("player_%d_item_3", i)] = p.Items.Item3.Name.GetName()
		r[fmt.Sprintf("player_%d_item_4", i)] = p.Items.Item4.Name.GetName()
		r[fmt.Sprintf("player_%d_item_5", i)] = p.Items.Item5.Name.GetName()
		r[fmt.Sprintf("player_%d_item_neutral", i)] = p.Items.ItemNeutral.Name.GetName()
		r[fmt.Sprintf("player_%d_backpack_0", i)] = p.Items.BackpackItem0.Name.GetName()
		r[fmt.Sprintf("player_%d_backpack_1", i)] = p.Items.BackpackItem1.Name.GetName()
		r[fmt.Sprintf("player_%d_backpack_2", i)] = p.Items.BackpackItem2.Name.GetName()
		i++
	})

	if details.Victory.RadiantWon() {
		r["radiant_win"] = "true"
		r["dire_win"] = "false"
	} else {
		r["radiant_win"] = "false"
		r["dire_win"] = "true"
	}
	r["radiant_score"] = fmt.Sprintf("%d", details.Score.RadiantScore)
	r["dire_score"] = fmt.Sprintf("%d", details.Score.DireScore)

	//time label
	if int64(details.Duration.Seconds())%60 > 10 {
		r["match_length"] = fmt.Sprintf("%d:%d", int64(details.Duration.Minutes()), int64(details.Duration.Seconds())%60)
	} else {
		r["match_length"] = fmt.Sprintf("%d:%d0", int64(details.Duration.Minutes()), int64(details.Duration.Seconds())%60)
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
