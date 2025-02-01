package objects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	mdl "github.com/AMETORY/whatsmeow-client/model"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type WaSession struct {
	Container *sqlstore.Container
	Ctx       context.Context
	Clients   []*whatsmeow.Client
	DB        *gorm.DB
}

func NewWaSession(ctx context.Context, container *sqlstore.Container, db *gorm.DB) *WaSession {
	return &WaSession{
		DB:        db,
		Container: container,
		Ctx:       ctx,
		Clients:   []*whatsmeow.Client{},
	}
}
func (ws *WaSession) AddSession(client *whatsmeow.Client) {
	ws.Clients = append(ws.Clients, client)
}

func (ws *WaSession) GetEventHandler(client *whatsmeow.Client, qrWait chan string) func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.PairSuccess:
			fmt.Println("Pair success", client.Store.ID)
			qrWait <- client.Store.ID.String()
		case *events.Connected:

			fmt.Println("Connected", client.Store.ID)
		case *events.Message:
			var WaDevice mdl.WaDevice
			err := ws.DB.Where("j_id = ?", client.Store.ID.String()).First(&WaDevice).Error
			if err != nil {
				fmt.Println(err)
			}
			var messageBody = v.Message.GetConversation()
			if messageBody == "ping" {
				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
					Conversation: proto.String("pong " + WaDevice.Webhook),
				})

			}
			if WaDevice.Webhook != "" {
				// LogJson(v.Message)
				b, _ := json.Marshal(v.Message)

				// fmt.Println(string(b))
				req, err := http.NewRequest("POST", WaDevice.Webhook, bytes.NewBuffer(b))
				if err != nil {
					fmt.Println(err)
				}
				req.Header.Set("Content-Type", "application/json")
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
				}
				defer resp.Body.Close()

			}
		}
	}
}

func LogJson(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
}
