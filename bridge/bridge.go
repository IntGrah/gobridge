package bridge

type Message struct {
	Reply       *Association
	Username    string
	Text        string
	Attachments []Attachment
}

type Attachment struct {
	Filename string
	Data     []byte
	MimeType string
}

type Association struct {
	DC  string
	WA  string
	JID string
}
