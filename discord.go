package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

func HandleDiscordMessage(s *discordgo.Session, dcMsg *discordgo.MessageCreate) {
	if dcMsg.ChannelID != dcChanID || dcMsg.Author.Bot {
		return
	}

	var (
		err      error
		stanzaID string
		waJID    string

		username      string = dcMsg.Author.Username
		text          string = dcMsg.Content
		formattedText string
	)

	if dcMsg.ReferencedMessage != nil {
		if stanzaID, waJID, err = GetWaFromDc(dcMsg.ReferencedMessage.ID); err == nil {
			formattedText = Format(username, text)
		} else {
			formattedText = FormatQuote(dcMsg.ReferencedMessage.Author.Username, dcMsg.ReferencedMessage.Content, username, text)
		}
	} else {
		formattedText = Format(username, text)
	}

	for embed := range dcMsg.Embeds {
		formattedText += "\nEmbed: " + dcMsg.Embeds[embed].URL
	}
	for attachment := range dcMsg.Attachments {
		formattedText += "\n" + dcMsg.Attachments[attachment].URL
	}

	waMessage := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &formattedText,
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    &stanzaID,
				Participant: &waJID,
			},
		},
	}

	// jpeg := "image/gif"

	// if len(dcMsg.Embeds) > 0 {
	// 	imageBytes, _ := DownloadDc(dcMsg.Embeds[0].URL)
	// 	uploadResp, err := waClient.Upload(context.Background(), imageBytes, whatsmeow.MediaImage)
	// 	if err != nil {
	// 		logIf(err)
	// 		return
	// 	}
	// 	waMessage = &waE2E.Message{
	// 		ImageMessage: &waE2E.ImageMessage{
	// 			URL:           &uploadResp.URL,
	// 			Mimetype:      &jpeg,
	// 			Caption:       &formattedText,
	// 			FileEncSHA256: uploadResp.FileEncSHA256,
	// 			FileSHA256:    uploadResp.FileSHA256,
	// 			FileLength:    &uploadResp.FileLength,
	// 			MediaKey:      uploadResp.MediaKey,
	// 			DirectPath:    &uploadResp.DirectPath,
	// 		},
	// 	}

	// 	fmt.Printf("waMessage: %v\n", waMessage)
	// }

	waResp, err := waClient.SendMessage(context.Background(), waGroupJID, waMessage)
	logIf(err)
	Associate(dcMsg.ID, waResp.ID, "")
}

// func DownloadDc(url string) ([]byte, error) {
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return body, nil
// }
