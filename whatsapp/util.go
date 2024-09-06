package whatsapp

import (
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func GetNameFromJID(senderJid types.JID) string {
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

func ExtractText(msg *waE2E.Message) string {
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
