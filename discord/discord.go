package discord

import (
	"bytes"

	"github.com/IntGrah/gobridge/bridge"
	"github.com/bwmarrin/discordgo"
)

var (
	Token     string
	Client    *discordgo.Session
	ChannelID string
)

func Receive(dcMsg *discordgo.MessageCreate) (bridge.Message, string) {
	message := bridge.Message{
		Username: dcMsg.Author.Username,
		Text:     dcMsg.Content,
	}

	if dcMsg.ReferencedMessage != nil {
		message.Reply, _ = bridge.Assoc.FromDc(dcMsg.ReferencedMessage.ID)
	}

	for _, embed := range dcMsg.Embeds {
		message.Text += "\n" + embed.URL
	}

	for _, attachment := range dcMsg.Attachments {
		message.Text += "\n" + attachment.URL
	}

	return message, dcMsg.ID
}

func Post(message bridge.Message) string {
	files := make([]*discordgo.File, len(message.Attachments))
	for i, attachment := range message.Attachments {
		files[i] = &discordgo.File{
			Name:        attachment.Filename,
			Reader:      bytes.NewReader(attachment.Data),
			ContentType: attachment.MimeType,
		}
	}

	messageSend := &discordgo.MessageSend{
		Content: bridge.Format(message.Username, message.Text),
		Files:   files,
	}

	if message.Reply != nil {
		messageSend.Reference = &discordgo.MessageReference{
			MessageID: message.Reply.DC,
			ChannelID: ChannelID,
			GuildID:   Client.State.Application.GuildID,
		}
	}

	dcMsg, _ := Client.ChannelMessageSendComplex(ChannelID, messageSend)

	return dcMsg.ID
}
