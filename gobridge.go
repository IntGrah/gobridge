package main

import (
	"context"
	"log"
	"os"

	"github.com/IntGrah/gobridge/database"
	"github.com/IntGrah/gobridge/discord"
	"github.com/IntGrah/gobridge/richtext"
	"github.com/IntGrah/gobridge/whatsapp"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func handleDiscordMessageCreate(_ *discordgo.Session, dcMsg *discordgo.MessageCreate) {
	if dcMsg.ChannelID != discord.ChannelID || dcMsg.Author == nil || dcMsg.Author.Bot {
		return
	}
	message, dcMsgID := discord.Receive(dcMsg)
	waMsgID, waJID := whatsapp.Post(message)
	database.Assoc.Put(database.Association{DC: dcMsgID, WA: waMsgID, JID: waJID})
}

func handleDiscordMessageDelete(_ *discordgo.Session, dcMsg *discordgo.MessageDelete) {
	if dcMsg.ChannelID != discord.ChannelID {
		return
	}

	association, err := database.Assoc.FromDc(dcMsg.ID)
	if err != nil {
		return
	}
	fromMe := true
	messageDelete := &waE2E.Message{
		ProtocolMessage: &waE2E.ProtocolMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: &whatsapp.GroupJIDStr,
				ID:        &association.WA,
				FromMe:    &fromMe,
			},
			Type: waE2E.ProtocolMessage_REVOKE.Enum(),
		},
	}
	whatsapp.Client.SendMessage(context.Background(), whatsapp.GroupJID, messageDelete)
	database.Assoc.Delete(association)
}

func handleDiscordMessageUpdate(_ *discordgo.Session, dcMsg *discordgo.MessageUpdate) {
	if dcMsg.ChannelID != discord.ChannelID || dcMsg.Author == nil || dcMsg.Author.Bot {
		return
	}
	association, err := database.Assoc.FromDc(dcMsg.ID)
	if err != nil {
		return
	}
	fromMe := true
	messageEdit := &waE2E.Message{
		ProtocolMessage: &waE2E.ProtocolMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: &whatsapp.GroupJIDStr,
				ID:        &association.WA,
				FromMe:    &fromMe,
			},
			EditedMessage: &waE2E.Message{
				Conversation: &dcMsg.Content,
			},
			Type: waE2E.ProtocolMessage_MESSAGE_EDIT.Enum(),
		},
	}
	whatsapp.Client.SendMessage(context.Background(), whatsapp.GroupJID, messageEdit)
}

func handleWhatsAppMessage(waMsg *events.Message) {
	if waMsg.Info.Chat != whatsapp.GroupJID {
		return
	}
	if waMsg.Message.ProtocolMessage != nil {
		prot := waMsg.Message.ProtocolMessage
		if prot.GetType() == waE2E.ProtocolMessage_REVOKE {
			association, err := database.Assoc.FromWa(prot.Key.GetID())
			if err != nil {
				return
			}
			discord.Client.ChannelMessageDelete(discord.ChannelID, association.DC)
			database.Assoc.Delete(association)
		} else if prot.GetType() == waE2E.ProtocolMessage_MESSAGE_EDIT {
			association, err := database.Assoc.FromWa(prot.Key.GetID())
			if err != nil {
				return
			}
			formattedText := richtext.Format(whatsapp.GetNameFromJID(waMsg.Info.Sender), whatsapp.ExtractText(prot.EditedMessage))
			discord.Client.ChannelMessageEdit(discord.ChannelID, association.DC, formattedText)
		}
		return
	}
	message, waMsgID, waJID := whatsapp.Receive(waMsg)
	dcMsgID := discord.Post(message)
	database.Assoc.Put(database.Association{DC: dcMsgID, WA: waMsgID, JID: waJID})
}

func EventHandler(evt interface{}) {
	switch e := evt.(type) {
	case *events.Message:
		handleWhatsAppMessage(e)
	}
}

func init() {
	godotenv.Load(".env")
	database.Assoc = database.NewMySQL()
	discord.Token = os.Getenv("DISCORD_TOKEN")
	discord.ChannelID = os.Getenv("DISCORD_CHANNEL_ID")

	whatsapp.GroupJIDStr = os.Getenv("WHATSAPP_GROUP_JID")
	whatsapp.GroupJID, _ = types.ParseJID(whatsapp.GroupJIDStr)
}

func main() {
	// Setup Discord session
	discord.Client, _ = discordgo.New("Bot " + discord.Token)
	discord.Client.AddHandler(handleDiscordMessageCreate)
	discord.Client.AddHandler(handleDiscordMessageUpdate)
	discord.Client.AddHandler(handleDiscordMessageDelete)
	discord.Client.Open()
	defer discord.Client.Close()

	// Setup WhatsApp client
	whatsapp.Client = whatsmeow.NewClient(whatsapp.GetDevice(), nil)
	whatsapp.Client.AddEventHandler(EventHandler)
	defer whatsapp.Client.Disconnect()
	if whatsapp.Client.Store.ID == nil {
		qrChan, _ := whatsapp.Client.GetQRChannel(context.Background())
		whatsapp.Client.Connect()
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			}
		}
	} else {
		whatsapp.Client.Connect()
	}

	log.Println("Connected")
	select {} // Block until interrupted
}
