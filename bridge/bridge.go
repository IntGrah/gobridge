package bridge

import "github.com/IntGrah/gobridge/database"

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
