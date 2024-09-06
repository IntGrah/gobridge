package main

import (
	"context"
	"log"
	"os"

	"github.com/IntGrah/gobridge/bridge"
	"github.com/IntGrah/gobridge/discord"
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

var fromMe = true

func discordHandleMessageCreate(_ *discordgo.Session, dcMsg *discordgo.MessageCreate) {
	if dcMsg.ChannelID != discord.ChannelID || dcMsg.Author == nil || dcMsg.Author.Bot {
		return
	}
	message, dcMsgID := discord.Receive(dcMsg)
	waMsgID, waMsgJID := whatsapp.Post(message)
	bridge.Assoc.Put(bridge.Association{DC: dcMsgID, WA: waMsgID, JID: waMsgJID})
}

func discordHandleMessageDelete(_ *discordgo.Session, dcMsg *discordgo.MessageDelete) {
	if dcMsg.ChannelID != discord.ChannelID {
		return
	}

	association, err := bridge.Assoc.FromDc(dcMsg.ID)
	if err != nil {
		return
	}
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
	bridge.Assoc.Delete(association)
}

func discordHandleMessageUpdate(_ *discordgo.Session, dcMsg *discordgo.MessageUpdate) {
	if dcMsg.ChannelID != discord.ChannelID || dcMsg.Author == nil || dcMsg.Author.Bot {
		return
	}
	association, err := bridge.Assoc.FromDc(dcMsg.ID)
	if err != nil {
		return
	}
	formattedText := bridge.Format(dcMsg.Author.Username, dcMsg.Content)
	messageEdit := &waE2E.Message{
		ProtocolMessage: &waE2E.ProtocolMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: &whatsapp.GroupJIDStr,
				ID:        &association.WA,
				FromMe:    &fromMe,
			},
			EditedMessage: &waE2E.Message{
				Conversation: &formattedText,
			},
			Type: waE2E.ProtocolMessage_MESSAGE_EDIT.Enum(),
		},
	}
	whatsapp.Client.SendMessage(context.Background(), whatsapp.GroupJID, messageEdit)
}

func whatsappHandleMessage(waMsg *events.Message) {
	if waMsg.Info.Chat != whatsapp.GroupJID {
		return
	}
	if waMsg.Message.ProtocolMessage != nil { // Edited or deleted message
		prot := waMsg.Message.ProtocolMessage
		if prot.GetType() == waE2E.ProtocolMessage_REVOKE {
			association, err := bridge.Assoc.FromWa(prot.Key.GetID())
			if err != nil {
				return
			}
			discord.Client.ChannelMessageDelete(discord.ChannelID, association.DC)
			bridge.Assoc.Delete(association)
		} else if prot.GetType() == waE2E.ProtocolMessage_MESSAGE_EDIT {
			association, err := bridge.Assoc.FromWa(prot.Key.GetID())
			if err != nil {
				return
			}
			formattedText := bridge.Format(whatsapp.GetNameFromJID(waMsg.Info.Sender), whatsapp.ExtractText(prot.EditedMessage))
			discord.Client.ChannelMessageEdit(discord.ChannelID, association.DC, formattedText)
		}
		return
	}
	message, waMsgID, waJID := whatsapp.Receive(waMsg)
	dcMsgID := discord.Post(message)
	bridge.Assoc.Put(bridge.Association{DC: dcMsgID, WA: waMsgID, JID: waJID})
}

func whatsappHandleEvent(evt interface{}) {
	switch e := evt.(type) {
	case *events.Message:
		whatsappHandleMessage(e)
	}
}

func init() {
	godotenv.Load(".env")
	bridge.Assoc = bridge.NewMySQL()
	discord.Token = os.Getenv("DISCORD_TOKEN")
	discord.ChannelID = os.Getenv("DISCORD_CHANNEL_ID")
	whatsapp.GroupJIDStr = os.Getenv("WHATSAPP_GROUP_JID")
	whatsapp.GroupJID, _ = types.ParseJID(whatsapp.GroupJIDStr)
}

func main() {
	// Setup Discord session
	discord.Client, _ = discordgo.New("Bot " + discord.Token)
	discord.Client.AddHandler(discordHandleMessageCreate)
	discord.Client.AddHandler(discordHandleMessageUpdate)
	discord.Client.AddHandler(discordHandleMessageDelete)
	discord.Client.Open()
	defer discord.Client.Close()

	// Setup WhatsApp client
	whatsapp.Client = whatsmeow.NewClient(whatsapp.GetDevice(), nil)
	whatsapp.Client.AddEventHandler(whatsappHandleEvent)
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
