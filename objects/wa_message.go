package objects

import "go.mau.fi/whatsmeow/proto/waE2E"

// "github.com/AMETORY/whatsmeow-client/model"

type WaMessage struct {
	JID             string                 `json:"jid"`
	Text            string                 `json:"text"`
	FileType        string                 `json:"file_type"`
	FileUrl         string                 `json:"file_url"`
	To              string                 `json:"to"`
	IsGroup         bool                   `json:"is_group"`
	RefID           *string                `json:"ref_id"`
	RefFrom         *string                `json:"ref_from"`
	RefText         *string                `json:"ref_text"`
	ChatPresence    string                 `json:"chat_presence"`
	EventMessage    *waE2E.EventMessage    `json:"event_message,omitempty"`
	LocationMessage *waE2E.LocationMessage `json:"location_message,omitempty"`
	ContactMessage  *waE2E.ContactMessage  `json:"contact_message,omitempty"`
}
