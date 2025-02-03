package model

type WhatsmeowContact struct {
	OurJid       string `json:"our_jid"`
	TheirJid     string `json:"their_jid"`
	FirstName    string `json:"first_name"`
	FullName     string `json:"full_name"`
	PushName     string `json:"push_name"`
	BusinessName string `json:"business_name"`
	PhoneNumber  string `json:"phone_number"`
}
