package main

import (
	"database/sql"
	"github.com/Neirpyc/dota2api"
	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	L *log.Logger

	//apis
	D  dota2api.Dota2
	DB *sql.DB
	DG *discordgo.Session

	//data
	Heroes dota2api.Heroes
	Items  dota2api.Items
	Config config
)

type config struct {
	Token        string `yaml:"Token"`
	MariaDb      string `yaml:"MariaDb"`
	RemoveImages bool   `yaml:"RemoveImages"`
	ForceReload  bool   `yaml:"ForceReload"`
}

func init() {
	//create a new logger
	L = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds|log.Ltime|log.Llongfile)

	//parse the configuration file
	configByte, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		L.Fatal(err)
	}
	err = yaml.Unmarshal(configByte, &Config)
	if err != nil {
		L.Fatal(err)
	}

	for _, path := range []string{"assets/tmp/", "assets/heroes/full/", "assets/heroes/lg/",
		"assets/heroes/vert/", "assets/heroes/sb/", "assets/items/lg"} {
		if err := os.MkdirAll(path, 0775); err != nil {
			if os.IsNotExist(err) {
				L.Fatal(err)
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(4)

	//seed rand
	go func(wg *sync.WaitGroup) {
		rand.Seed(time.Now().Unix())
		wg.Done()
	}(&wg)

	//connect to the dota api
	go func(wg *sync.WaitGroup) {
		var err error
		D, err = dota2api.LoadConfig("config.yaml")
		if err != nil {
			L.Fatal(err)
		}

		Heroes, err = D.GetHeroes()
		if err != nil {
			L.Fatal(err)
		}

		//download images of heroes and items
		var wg0 sync.WaitGroup
		wg0.Add(2)
		go func(wg *sync.WaitGroup) {
			createHeroesImagesList()
			wg.Done()
		}(&wg0)
		go func(wg *sync.WaitGroup) {
			createItemsImagesList()
			wg.Done()
		}(&wg0)
		wg0.Wait()
		wg.Done()
	}(&wg)

	//connect to database
	go func(wg *sync.WaitGroup) {
		var err error
		DB, err = sql.Open("mysql", Config.MariaDb)
		if err != nil {
			L.Fatal(err)
		}

		_, err = DB.Query("USE dota_bot_discord")
		if err != nil {
			L.Println("Cannot find the 'dota_bot_discord' database: it will be created.")
			rows, err := DB.Query("CREATE DATABASE dota_bot_discord")
			if err != nil {
				L.Fatal(err)
			}
			err = rows.Close()
			if err != nil {
				L.Fatal(err)
			}

			L.Println("Created the 'dota_bot_discord' database successfully!")

			_, err = DB.Exec("USE dota_bot_discord")
			if err != nil {
				L.Fatal(err)
			}
			_, err = DB.Exec(`create table steam_id
		(
			discord_id BIGINT PRIMARY KEY not null,
			steam_id BIGINT not null
		);
		`)
			if err != nil {
				L.Fatal(err)
			}
			_, err = DB.Exec(`create unique index steam_ide_discord_id_uindex
		on steam_id (discord_id);`)
			if err != nil {
				L.Fatal(err)
			}
		}
		wg.Done()
	}(&wg)

	//open discord API
	go func(wg *sync.WaitGroup) {
		var err error
		DG, err = discordgo.New("Bot " + Config.Token)
		if err != nil {
			L.Fatal(err)
		}
		wg.Done()
	}(&wg)

	wg.Wait()
}

func main() {
	DG.AddHandler(HandleMessage)

	err := DG.Open()
	if err != nil {
		L.Fatal(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if err = os.RemoveAll("assets/tmp/"); err != nil {
		L.Println(err)
	}

	if Config.RemoveImages {
		if err = os.RemoveAll("assets/heroes/"); err != nil {
			L.Println(err)
		}
	}

	if Config.RemoveImages {
		if err = os.RemoveAll("assets/items/"); err != nil {
			L.Println(err)
		}
	}

	err = DG.Close()
	if err != nil {
		L.Fatal(err)
	}
}
