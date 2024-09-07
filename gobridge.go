package main

import (
	"context"
	"log"
	"os"

	"github.com/IntGrah/gobridge/database"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

var (
	fromMe           = true // constant
	discordToken     string
	discordClient    *discordgo.Session
	discordChannelID string
)

func whatsappHandleEvent(evt interface{}) {
	switch e := evt.(type) {
	case *events.Message:
		whatsappHandleMessage(e)
	}
}

func init() {
	godotenv.Load(".env")
	database.Assoc = database.NewMySQL()
	discordToken = os.Getenv("DISCORD_TOKEN")
	discordChannelID = os.Getenv("DISCORD_CHANNEL_ID")
	GroupJIDStr = os.Getenv("WHATSAPP_GROUP_JID")
	GroupJID, _ = types.ParseJID(GroupJIDStr)
}

func main() {
	// Setup Discord session
	discordClient, _ = discordgo.New("Bot " + discordToken)
	discordClient.AddHandler(discordHandleMessageCreate)
	discordClient.AddHandler(discordHandleMessageUpdate)
	discordClient.AddHandler(discordHandleMessageDelete)
	discordClient.Open()
	defer discordClient.Close()

	// Setup WhatsApp client
	storeContainer, _ := sqlstore.New("sqlite3", "file:session.db?_pragma=foreign_keys(1)&_pragma=busy_timeout=10000", nil)
	device, _ = storeContainer.GetFirstDevice()
	whatsAppClient = whatsmeow.NewClient(device, nil)
	whatsAppClient.AddEventHandler(whatsappHandleEvent)
	defer whatsAppClient.Disconnect()
	if whatsAppClient.Store.ID == nil {
		qrChan, _ := whatsAppClient.GetQRChannel(context.Background())
		whatsAppClient.Connect()
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			}
		}
	} else {
		whatsAppClient.Connect()
	}

	log.Println("Connected")
	select {} // Block until interrupted
}
