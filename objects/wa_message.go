package objects

type WaMessage struct {
	JID   string `json:"jid"`
	Text  string `json:"text"`
	File  string `json:"file"`
	Image string `json:"image"`
	To    string `json:"to"`
}
