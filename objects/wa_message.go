package objects

type WaMessage struct {
	JID  string `json:"jid"`
	Text string `json:"text"`
	To   string `json:"to"`
}
