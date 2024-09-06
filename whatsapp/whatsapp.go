package whatsapp

import (
	"context"
	"time"

	"github.com/IntGrah/gobridge/bridge"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

var (
	Client      *whatsmeow.Client
	GroupJIDStr string
	GroupJID    types.JID
	waContacts  = make(map[types.JID]types.ContactInfo)
	device      *store.Device
)

func GetDevice() *store.Device {
	storeContainer, _ := sqlstore.New("sqlite3", "file:session.db?_pragma=foreign_keys(1)&_pragma=busy_timeout=10000", nil)
	device, _ = storeContainer.GetFirstDevice()
	return device
}

func Receive(waMsg *events.Message) (bridge.Message, string, string) {
	message := bridge.Message{
		Username: GetNameFromJID(waMsg.Info.Sender),
		Text:     ExtractText(waMsg.Message),
	}

	msg := waMsg.Message
	if ci := getContextInfo(msg); ci != nil {
		message.Reply, _ = bridge.Assoc.FromWa(ci.GetStanzaID())
	}
	if msg.ProtocolMessage != nil && msg.ProtocolMessage.GetType() == waE2E.ProtocolMessage_MESSAGE_EDIT {
		message.Reply, _ = bridge.Assoc.FromWa(waMsg.Message.ProtocolMessage.Key.GetID())
	}

	var attachment bridge.Attachment
	var mediaMessage interface {
		whatsmeow.DownloadableMessage
		GetMimetype() string
	}

	switch {
	case msg.ImageMessage != nil:
		mediaMessage = msg.ImageMessage
	case msg.VideoMessage != nil:
		mediaMessage = msg.VideoMessage
	case msg.AudioMessage != nil:
		mediaMessage = msg.AudioMessage
	case msg.DocumentMessage != nil:
		mediaMessage = msg.DocumentMessage
	}

	if mediaMessage != nil {
		attachment.Data, _ = Client.Download(mediaMessage)
		attachment.MimeType = mediaMessage.GetMimetype()
		ext := bridge.MimeTypeToExtension(attachment.MimeType)
		attachment.Filename = "MM-" + time.Now().UTC().Format("20060102") + "-WA0000" + ext

		message.Attachments = append(message.Attachments, attachment)
	}
	return message, waMsg.Info.ID, waMsg.Info.Sender.String()
}

func Post(message bridge.Message) (string, string) {
	formattedText := bridge.Format(message.Username, message.Text)

	waMessage := &waE2E.Message{}

	if message.Reply != nil {
		waMessage.ExtendedTextMessage = &waE2E.ExtendedTextMessage{
			Text: &formattedText,
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    &message.Reply.WA,
				Participant: &message.Reply.JID,
			},
		}
	} else {
		waMessage.Conversation = &formattedText
	}

	// TODO
	// for _, attachment := range message.Attachments {
	// 	jpeg := "image/jpeg"
	// 	uploadResp, err := Client.Upload(context.Background(), attachment.Data, whatsmeow.MediaImage)
	// 	if err != nil {
	// 		fmt.Printf("Error uploading attachment: %v\n", err)
	// 		continue
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

	waResp, _ := Client.SendMessage(context.Background(), GroupJID, waMessage)
	return waResp.ID, Client.Store.ID.String()
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
