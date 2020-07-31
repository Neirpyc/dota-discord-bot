package main

import (
	"bytes"
	"fmt"
	"github.com/l2x/dota2api"
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

func (h HeroList) getName(id int) string {
	for i := 0; i < len(h); i++ {
		if id == h[i].ID {
			return h[i].Name
		}
	}
	return ""
}

func getMatchImgSmall(match dota2api.MatchSummary, steamId string) image.Image {
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
	r["match_length"] = fmt.Sprintf("%d:%d", int64(d.Minutes()), int64(d.Seconds())%60)

	path, err := r.applyTemplate("assets/templates/small.html")
	if err != nil {
		L.Println(err)
		return nil
	}
	imgData := screenshotFile(path)
	if err = os.Remove("asssets/" + path); err != nil {
		L.Println(err)
	}
	if img, err := png.Decode(bytes.NewReader(imgData)); err != nil {
		L.Fatal(err)
		return nil
	} else {
		return img
	}
}

const url = "http://cdn.dota2.com/apps/dota2/images/heroes/<name>_<size>.<ext>"

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
			customUrl := strings.Replace(url, "<name>", name, 1)

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
