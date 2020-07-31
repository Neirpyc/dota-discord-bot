package main

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/l2x/dota2api"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
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
	Heroes []dota2api.Hero
	Config config
)

type config struct {
	Token        string `yaml:"Token"`
	MariaDb      string `yaml:"MariaDb"`
	RemoveHeroes bool   `yaml:"RemoveHeroes"`
	ForceReload  bool   `yaml:"ForceReload"`
}

func init() {
	//create a new logger
	L = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds|log.Ltime|log.Llongfile)

	//seed rand
	rand.Seed(time.Now().Unix())

	//parse the configuration file
	configByte, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		L.Fatal(err)
	}
	err = yaml.Unmarshal(configByte, &Config)
	if err != nil {
		L.Fatal(err)
	}

	//connect to the dota api
	D, err = dota2api.LoadConfig("config.ini")
	if err != nil {
		L.Fatal(err)
	}

	Heroes, err = D.GetHeroes()
	if err != nil {
		L.Fatal(err)
	}

	//connect to database
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

	//open discord API
	DG, err = discordgo.New("Bot " + Config.Token)
	if err != nil {
		L.Fatal(err)
	}

	//create the tmp dir
	for _, path := range []string{"assets/tmp/", "assets/heroes/full/", "assets/heroes/lg/",
		"assets/heroes/vert/", "assets/heroes/sb/"} {
		if err = os.MkdirAll(path, 0775); err != nil {
			if os.IsNotExist(err) {
				L.Fatal(err)
			}
		}
	}

	//download the heads of all heroes
	createHeroesImagesList()
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

	if Config.RemoveHeroes {
		if err = os.RemoveAll("assets/heroes/"); err != nil {
			L.Println(err)
		}
	}

	err = DG.Close()
	if err != nil {
		L.Fatal(err)
	}
}
