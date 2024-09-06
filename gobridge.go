package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
)

var (
	db            *sql.DB
	dcToken       string
	dcBot         *discordgo.Session
	waClient      *whatsmeow.Client
	dcChanID      string
	waGroupJIDStr string
	waGroupJID, _ = types.ParseJID(waGroupJIDStr)
	waContacts    = make(map[types.JID]types.ContactInfo)
	device        *store.Device
)

func getDevice() *store.Device {
	storeContainer, _ := sqlstore.New("sqlite3", "file:session.db?_pragma=foreign_keys(1)&_pragma=busy_timeout=10000", nil)
	device, _ = storeContainer.GetFirstDevice()
	return device
}

func main() {
	var err error
	godotenv.Load(".env")
	dcToken = os.Getenv("DISCORD_TOKEN")
	dcChanID = os.Getenv("DISCORD_CHANNEL_ID")
	waGroupJIDStr = os.Getenv("WHATSAPP_GROUP_JID")

	// Connect to MySQL database
	cfg := mysql.Config{
		User:   os.Getenv("MYSQL_USER"),
		Passwd: os.Getenv("MYSQL_PASSWORD"),
		Net:    "tcp",
		Addr:   os.Getenv("MYSQL_HOST"),
		DBName: os.Getenv("MYSQL_DATABASE"),
	}
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("select * from assoc")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer rows.Close()

	// Setup Discord session
	dcBot, _ = discordgo.New("Bot " + dcToken)
	dcBot.AddHandler(HandleDiscordMessage)
	dcBot.Open()
	defer dcBot.Close()

	// Setup WhatsApp client
	deviceStore := getDevice()

	waClient = whatsmeow.NewClient(deviceStore, nil)
	waClient.AddEventHandler(eventHandler)
	defer waClient.Disconnect()

	if waClient.Store.ID == nil {
		qrChan, _ := waClient.GetQRChannel(context.Background())
		waClient.Connect()
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				log.Println("Login event:", evt.Event)
			}
		}
	} else {
		waClient.Connect()
	}

	log.Println("Connected")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop // Block until signal is received
}

func logIf(err error) {
	if err != nil {
		log.Println(err)
	}
}
