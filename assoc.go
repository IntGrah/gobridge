package main

import (
	"fmt"
	"log"

	"go.mau.fi/whatsmeow/types"
)

func GetWaFromDc(dc string) (string, string, error) {
	var (
		id      int
		dcMsgID string
		waMsgID string
		waJID   string
	)
	rows := db.QueryRow("SELECT * FROM assoc WHERE dc = ?", dc)
	if err := rows.Scan(&id, &dcMsgID, &waMsgID, &waJID); err != nil {
		return "", "", fmt.Errorf("failed to scan row: %v", err)
	}
	return waMsgID, waJID, nil
}

func GetDcfromWa(wa string) (string, error) {
	var (
		id      int
		dcMsgID string
		waMsgID string
		waJID   string
	)
	rows := db.QueryRow("SELECT * FROM assoc WHERE wa = ?", wa)
	if err := rows.Scan(&id, &dcMsgID, &waMsgID, &waJID); err != nil {
		return "", fmt.Errorf("failed to scan row: %v", err)
	}
	return dcMsgID, nil
}

func Associate(dc, wa, jid string) error {
	log.Println("Associating", dc, wa, jid)
	if _, err := db.Exec("INSERT INTO assoc (dc, wa, jid) VALUES (?, ?, ?)", dc, wa, jid); err != nil {
		return fmt.Errorf("failed to insert into assoc: %v", err)
	}
	return nil
}

func GetNameFromJID(senderJid types.JID) string {
	contact, exists := waContacts[senderJid]
	if !exists {
		waC, err := device.Contacts.GetAllContacts()
		logIf(err)
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
