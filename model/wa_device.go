package model

type WaDevice struct {
	Session string `json:"session"`
	JID     string `json:"jid"`
	Webhook string `json:"webhook"`
}
