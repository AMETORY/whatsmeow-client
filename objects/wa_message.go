package objects

type WaMessage struct {
	JID      string `json:"jid"`
	Text     string `json:"text"`
	FileType string `json:"file_type"`
	FileUrl  string `json:"file_url"`
	To       string `json:"to"`
}
