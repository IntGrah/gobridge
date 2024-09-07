package main

import (
	"context"

	"github.com/IntGrah/gobridge/database"
	"github.com/bwmarrin/discordgo"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

func discordHandleMessageCreate(_ *discordgo.Session, dcMsg *discordgo.MessageCreate) {
	if dcMsg.ChannelID != discordChannelID || dcMsg.Author == nil || dcMsg.Author.Bot {
		return
	}

	messageText := dcMsg.Content
	// TODO upload Discord attachments to WhatsApp servers
	for _, embed := range dcMsg.Embeds {
		messageText += "\n" + embed.URL
	}
	for _, attachment := range dcMsg.Attachments {
		messageText += "\n" + attachment.URL
	}
	formattedText := quoteFormat(dcMsg.Author.Username, messageText)

	waMessage := &waE2E.Message{}
	if dcMsg.ReferencedMessage != nil {
		messageReply, _ := database.Assoc.FromDc(dcMsg.ReferencedMessage.ID)
		if messageReply != nil {
			waMessage.ExtendedTextMessage = &waE2E.ExtendedTextMessage{
				Text: &formattedText,
				ContextInfo: &waE2E.ContextInfo{
					StanzaID:    &messageReply.WA,
					Participant: &messageReply.JID,
				},
			}
		} else {
			waMessage.Conversation = &formattedText
		}
	} else {
		waMessage.Conversation = &formattedText
	}

	waResp, _ := whatsAppClient.SendMessage(context.Background(), GroupJID, waMessage)
	database.Assoc.Put(database.Association{DC: dcMsg.ID, WA: waResp.ID, JID: whatsAppClient.Store.ID.String()})
}

func discordHandleMessageDelete(_ *discordgo.Session, dcMsg *discordgo.MessageDelete) {
	if dcMsg.ChannelID != discordChannelID {
		return
	}

	association, err := database.Assoc.FromDc(dcMsg.ID)
	if err != nil {
		return
	}
	messageDelete := &waE2E.Message{
		ProtocolMessage: &waE2E.ProtocolMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: &GroupJIDStr,
				ID:        &association.WA,
				FromMe:    &fromMe,
			},
			Type: waE2E.ProtocolMessage_REVOKE.Enum(),
		},
	}
	whatsAppClient.SendMessage(context.Background(), GroupJID, messageDelete)
	database.Assoc.Delete(association)
}

func discordHandleMessageUpdate(_ *discordgo.Session, dcMsg *discordgo.MessageUpdate) {
	if dcMsg.ChannelID != discordChannelID || dcMsg.Author == nil || dcMsg.Author.Bot {
		return
	}
	association, err := database.Assoc.FromDc(dcMsg.ID)
	if err != nil {
		return
	}
	formattedText := quoteFormat(dcMsg.Author.Username, dcMsg.Content)
	messageEdit := &waE2E.Message{
		ProtocolMessage: &waE2E.ProtocolMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: &GroupJIDStr,
				ID:        &association.WA,
				FromMe:    &fromMe,
			},
			EditedMessage: &waE2E.Message{
				Conversation: &formattedText,
			},
			Type: waE2E.ProtocolMessage_MESSAGE_EDIT.Enum(),
		},
	}
	whatsAppClient.SendMessage(context.Background(), GroupJID, messageEdit)
}
