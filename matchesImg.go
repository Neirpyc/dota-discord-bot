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

func getMatchReplacement(match dota2api.MatchSummary, steamId string) *Replacement {
	r := sync.Map{}

	//getting match details
	details, err := D.GetMatchDetails(dota2api.MatchId(match.MatchId))
	if err != nil {
		L.Fatal(err)
	}

	wait := details.GoForEachPlayerI(func(p dota2api.PlayerDetails, index int) {
		r.Store(fmt.Sprintf("hero_name_%d", index), p.Hero.Name.GetName())
		if fmt.Sprintf("%d", p.AccountId) == steamId {
			r.Store(fmt.Sprintf("class_hero_%d", index), "player")
		} else {
			r.Store(fmt.Sprintf("class_hero_%d", index), "")
		}

		r.Store(fmt.Sprintf("kills_player_%d", index), fmt.Sprintf("%d", p.KDA.Kills))
		r.Store(fmt.Sprintf("assists_player_%d", index), fmt.Sprintf("%d", p.KDA.Assists))
		r.Store(fmt.Sprintf("deaths_player_%d", index), fmt.Sprintf("%d", p.KDA.Deaths))

		//gold
		r.Store(fmt.Sprintf("gold_player_%d", index), p.Stats.Gold.NetWorth().ToString())

		//items
		r.Store(fmt.Sprintf("player_%d_item_0", index), p.Items.Item0.Name.GetName())
		r.Store(fmt.Sprintf("player_%d_item_1", index), p.Items.Item1.Name.GetName())
		r.Store(fmt.Sprintf("player_%d_item_2", index), p.Items.Item2.Name.GetName())
		r.Store(fmt.Sprintf("player_%d_item_3", index), p.Items.Item3.Name.GetName())
		r.Store(fmt.Sprintf("player_%d_item_4", index), p.Items.Item4.Name.GetName())
		r.Store(fmt.Sprintf("player_%d_item_5", index), p.Items.Item5.Name.GetName())
		r.Store(fmt.Sprintf("player_%d_item_neutral", index), p.Items.ItemNeutral.Name.GetName())
		r.Store(fmt.Sprintf("player_%d_backpack_0", index), p.Items.BackpackItem0.Name.GetName())
		r.Store(fmt.Sprintf("player_%d_backpack_1", index), p.Items.BackpackItem1.Name.GetName())
		r.Store(fmt.Sprintf("player_%d_backpack_2", index), p.Items.BackpackItem2.Name.GetName())
	})

	if details.Victory.RadiantWon() {
		r.Store("radiant_win", "true")
		r.Store("dire_win", "false")
	} else {
		r.Store("radiant_win", "false")
		r.Store("dire_win", "true")
	}
	r.Store("radiant_score", fmt.Sprintf("%d", details.Score.RadiantScore))
	r.Store("dire_score", fmt.Sprintf("%d", details.Score.DireScore))

	//time label
	if int64(details.Duration.Seconds())%60 > 10 {
		r.Store("match_length", fmt.Sprintf("%d:%d", int64(details.Duration.Minutes()), int64(details.Duration.Seconds())%60))
	} else {
		r.Store("match_length", fmt.Sprintf("%d:%d0", int64(details.Duration.Minutes()), int64(details.Duration.Seconds())%60))
	}

	//game date
	r.Store("game_date", fmt.Sprintf("%2.2d/%2.2d/%4.4d", match.StartTime.Month(), match.StartTime.Day(), match.StartTime.Year()))
	r.Store("game_type", match.LobbyType.GetName())

	//items

	wait()

	return (*Replacement)(&r)
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
	Heroes.GoForEach(func(hero dota2api.Hero) {
		sizes := []string{"lg", "sb", "full", "vert"}
		for i, size := range sizes {

			path := "assets/heroes/" + size + "/" + hero.Name.GetName() + ".png"
			if _, err := os.Stat(path); !os.IsNotExist(err) && !Config.ForceReload {
				continue
			}

			img, err0 := D.GetHeroImage(hero, dota2api.HeroImageSize(i))
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
	})()
}

func createItemsImagesList() {
	Items.GoForEach(func(item dota2api.Item) {
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
	})()
}
