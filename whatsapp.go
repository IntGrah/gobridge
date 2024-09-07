package main

import (
	"bytes"
	"time"

	"github.com/IntGrah/gobridge/database"
	"github.com/bwmarrin/discordgo"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type Message struct {
	Reply       *database.Association
	Username    string
	Text        string
	Attachments []Attachment
}

type Attachment struct {
	Filename string
	Data     []byte
	MimeType string
}

var (
	whatsAppClient *whatsmeow.Client
	GroupJIDStr    string
	GroupJID       types.JID
	waContacts     = make(map[types.JID]types.ContactInfo)
	device         *store.Device
)

func whatsappHandleMessage(waMsg *events.Message) {
	if waMsg.Info.Chat != GroupJID {
		return
	}
	switch {
	case waMsg.Message.Conversation != nil:
		fallthrough
	case waMsg.Message.ExtendedTextMessage != nil:
		fallthrough
	case waMsg.Message.ImageMessage != nil:
		fallthrough
	case waMsg.Message.VideoMessage != nil:
		fallthrough
	case waMsg.Message.AudioMessage != nil:
		fallthrough
	case waMsg.Message.DocumentMessage != nil:
		message, waMsgID, waJID := normaliseWhatsAppMessage(waMsg)
		sendToDiscord(message, waMsgID, waJID)
	case waMsg.Message.ProtocolMessage != nil && waMsg.Message.ProtocolMessage.GetType() == waE2E.ProtocolMessage_REVOKE:
		association, err := database.Assoc.FromWa(waMsg.Message.ProtocolMessage.Key.GetID())
		if err != nil {
			return
		}
		discordClient.ChannelMessageDelete(discordChannelID, association.DC)
		database.Assoc.Delete(association)
	case waMsg.Message.ProtocolMessage != nil && waMsg.Message.ProtocolMessage.GetType() == waE2E.ProtocolMessage_MESSAGE_EDIT:
		association, err := database.Assoc.FromWa(waMsg.Message.ProtocolMessage.Key.GetID())
		if err != nil {
			return
		}
		formattedText := quoteFormat(getNameFromJID(waMsg.Info.Sender), extractText(waMsg.Message.ProtocolMessage.EditedMessage))
		discordClient.ChannelMessageEdit(discordChannelID, association.DC, formattedText)
	}
}

func sendToDiscord(message Message, waMsgID, waJID string) {
	files := make([]*discordgo.File, len(message.Attachments))
	for i, attachment := range message.Attachments {
		files[i] = &discordgo.File{
			Name:        attachment.Filename,
			Reader:      bytes.NewReader(attachment.Data),
			ContentType: attachment.MimeType,
		}
	}

	messageSend := &discordgo.MessageSend{
		Content: quoteFormat(message.Username, message.Text),
		Files:   files,
	}

	if message.Reply != nil {
		messageSend.Reference = &discordgo.MessageReference{
			MessageID: message.Reply.DC,
			ChannelID: discordChannelID,
			GuildID:   discordClient.State.Application.GuildID,
		}
	}

	dcMsg, _ := discordClient.ChannelMessageSendComplex(discordChannelID, messageSend)
	database.Assoc.Put(database.Association{DC: dcMsg.ID, WA: waMsgID, JID: waJID})
}

func normaliseWhatsAppMessage(waMsg *events.Message) (Message, string, string) {
	message := Message{
		Username: getNameFromJID(waMsg.Info.Sender),
		Text:     extractText(waMsg.Message),
	}

	if ci := getContextInfo(waMsg.Message); ci != nil {
		message.Reply, _ = database.Assoc.FromWa(ci.GetStanzaID())
	}
	if waMsg.Message.ProtocolMessage != nil && waMsg.Message.ProtocolMessage.GetType() == waE2E.ProtocolMessage_MESSAGE_EDIT {
		message.Reply, _ = database.Assoc.FromWa(waMsg.Message.ProtocolMessage.Key.GetID())
	}

	var attachment Attachment
	var mediaMessage interface {
		whatsmeow.DownloadableMessage
		GetMimetype() string
	}

	switch {
	case waMsg.Message.ImageMessage != nil:
		mediaMessage = waMsg.Message.ImageMessage
	case waMsg.Message.VideoMessage != nil:
		mediaMessage = waMsg.Message.VideoMessage
	case waMsg.Message.AudioMessage != nil:
		mediaMessage = waMsg.Message.AudioMessage
	case waMsg.Message.DocumentMessage != nil:
		mediaMessage = waMsg.Message.DocumentMessage
	}

	if mediaMessage != nil {
		attachment.Data, _ = whatsAppClient.Download(mediaMessage)
		attachment.MimeType = mediaMessage.GetMimetype()
		ext := mimeTypeToExtension(attachment.MimeType)
		attachment.Filename = "MM-" + time.Now().UTC().Format("20060102") + "-WA0000" + ext

		message.Attachments = append(message.Attachments, attachment)
	}
	return message, waMsg.Info.ID, waMsg.Info.Sender.String()
}

func getContextInfo(msg *waE2E.Message) *waE2E.ContextInfo {
	switch {
	case msg.Conversation != nil:
		return nil
	case msg.ExtendedTextMessage != nil:
		return msg.ExtendedTextMessage.GetContextInfo()
	case msg.ImageMessage != nil:
		return msg.ImageMessage.GetContextInfo()
	case msg.VideoMessage != nil:
		return msg.VideoMessage.GetContextInfo()
	case msg.DocumentMessage != nil:
		return msg.DocumentMessage.GetContextInfo()
	default:
		return nil
	}
}

func getNameFromJID(senderJid types.JID) string {
	contact, exists := waContacts[senderJid]
	if !exists {
		waC, _ := device.Contacts.GetAllContacts()
		contact = waC[senderJid]
		waContacts = waC
	}
	if contact.FullName != "" {
		return contact.FullName
	}
	if contact.FirstName != "" {
		return contact.FirstName
	}
	if contact.PushName != "" {
		return contact.PushName
	}
	return "A WhatsApp user"
}

func extractText(msg *waE2E.Message) string {
	switch {
	case msg.Conversation != nil:
		return msg.GetConversation()
	case msg.ExtendedTextMessage != nil:
		return msg.ExtendedTextMessage.GetText()
	case msg.ImageMessage != nil:
		return msg.ImageMessage.GetCaption()
	case msg.VideoMessage != nil:
		return msg.VideoMessage.GetCaption()
	case msg.DocumentMessage != nil:
		return msg.DocumentMessage.GetCaption()
	default:
		return ""
	}
}
