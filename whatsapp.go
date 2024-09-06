package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func eventHandler(evt interface{}) {
	switch e := evt.(type) {
	case *events.Message:
		handleWhatsAppMessage(e)
	case *events.GroupInfo:
		return
	}
}

func handleWhatsAppMessage(waMsg *events.Message) {
	if waMsg.Info.Chat != waGroupJID {
		return
	}

	var (
		dcMsg         *discordgo.Message
		dcMsgRefReply *discordgo.MessageReference
		msg           *waE2E.Message     = waMsg.Message
		ci            *waE2E.ContextInfo = getContextInfo(msg)
		isReply       bool               = ci != nil

		filename string
		data     []byte
		mimeType string
		ext      string

		qUsername string
		qText     string
		username  string = GetNameFromJID(waMsg.Info.Sender)
		text      string = extractText(msg)

		messageSend   *discordgo.MessageSend
		formattedText string = Format(username, text)
	)

	if isReply {
		qJID, _ := types.ParseJID(ci.GetParticipant())
		qUsername = GetNameFromJID(qJID)
		if ci.QuotedMessage != nil {
			qText = ci.QuotedMessage.GetConversation()
		}

		if dcMsgIdReply, err := GetDcfromWa(ci.GetStanzaID()); err != nil {
			formattedText = FormatQuote(qUsername, qText, username, text)
		} else {
			dcMsgRefReply = getDcMsgRef(dcMsgIdReply)
		}
	}

	switch {
	case msg.ImageMessage != nil:
		data, _ = waClient.Download(msg.ImageMessage)
		mimeType = msg.ImageMessage.GetMimetype()
		ext = MimeTypeToExtension(mimeType)
		fmt.Printf("mimeType: %v\n", mimeType)
		fmt.Printf("ext: %v\n", ext)
		filename = "image" + ext
	case msg.VideoMessage != nil:
		data, _ = waClient.Download(msg.VideoMessage)
		mimeType = msg.VideoMessage.GetMimetype()
		ext = MimeTypeToExtension(mimeType)
		fmt.Printf("mimeType: %v\n", mimeType)
		fmt.Printf("ext: %v\n", ext)
		filename = "video" + ext
	case msg.AudioMessage != nil:
		data, _ = waClient.Download(msg.AudioMessage)
		mimeType = msg.AudioMessage.GetMimetype()
		ext = MimeTypeToExtension(mimeType)
		fmt.Printf("mimeType: %v\n", mimeType)
		fmt.Printf("ext: %v\n", ext)
		filename = "audio" + ext
	case msg.DocumentMessage != nil:
		data, _ = waClient.Download(msg.DocumentMessage)
		mimeType = msg.DocumentMessage.GetMimetype()
		ext = MimeTypeToExtension(mimeType)
		fmt.Printf("mimeType: %v\n", mimeType)
		fmt.Printf("ext: %v\n", ext)
		filename = "document" + ext
	}

	switch {
	case msg.Conversation != nil:
		log.Printf("Conversation: %s", text)
		dcMsg, _ = dcBot.ChannelMessageSend(dcChanID, formattedText)

	case msg.ExtendedTextMessage != nil:
		log.Printf("Extended text: %s", text)
		messageSend = &discordgo.MessageSend{
			Content:   formattedText,
			Reference: dcMsgRefReply,
		}
		dcMsg, _ = dcBot.ChannelMessageSendComplex(dcChanID, messageSend)

	case msg.ImageMessage != nil:
		fallthrough
	case msg.VideoMessage != nil:
		fallthrough
	case msg.AudioMessage != nil:
		fallthrough
	case msg.DocumentMessage != nil:
		messageSend = &discordgo.MessageSend{
			Reference: dcMsgRefReply,
			Content:   formattedText,
			Files: []*discordgo.File{
				{
					Name:        filename,
					Reader:      bytes.NewReader(data),
					ContentType: mimeType,
				},
			},
		}
		dcMsg, _ = dcBot.ChannelMessageSendComplex(dcChanID, messageSend)

	case msg.ProtocolMessage != nil && msg.ProtocolMessage.GetType() == 0:
		log.Printf("Deletion message")

		if dcMsgID, err := GetDcfromWa(msg.ProtocolMessage.Key.GetID()); err != nil {
			err = dcBot.ChannelMessageDelete(dcChanID, dcMsgID)
			logIf(err)
		}

	default:
		log.Println("Unknown message type")
		log.Println(waMsg.RawMessage)
	}

	if dcMsg != nil {
		Associate(dcMsg.ID, waMsg.Info.ID, waMsg.Info.Sender.String())
	}
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
	case msg.ProtocolMessage != nil && msg.ProtocolMessage.GetType() == 0:
		return ""
	default:
		return ""
	}
}

func getContextInfo(msg *waE2E.Message) *waE2E.ContextInfo {
	switch {
	case msg.Conversation != nil: // Text
		return nil
	case msg.ExtendedTextMessage != nil: // Text, reply
		return msg.ExtendedTextMessage.GetContextInfo()
	case msg.ImageMessage != nil: // Text, image, reply
		return msg.ImageMessage.GetContextInfo()
	case msg.VideoMessage != nil:
		return msg.VideoMessage.GetContextInfo()
	case msg.DocumentMessage != nil:
		return msg.DocumentMessage.GetContextInfo()
	case msg.ProtocolMessage != nil && msg.ProtocolMessage.GetType() == 0:
		return nil
	default:
		return nil
	}
}

func getDcMsgRef(msgId string) *discordgo.MessageReference {
	return &discordgo.MessageReference{
		MessageID: msgId,
		ChannelID: dcChanID,
		GuildID:   dcBot.State.Application.GuildID,
	}
}
