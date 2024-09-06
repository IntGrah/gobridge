package whatsapp

import "go.mau.fi/whatsmeow/types"

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
