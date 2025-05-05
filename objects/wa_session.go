package objects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	mdl "github.com/AMETORY/whatsmeow-client/model"
	"github.com/AMETORY/whatsmeow-client/utils"

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
	var WaDevice mdl.WaDevice
	if client.Store.ID != nil {
		fmt.Println("client store", client.Store.ID.String())
		err := ws.DB.Where("j_id = ?", client.Store.ID.String()).First(&WaDevice).Error
		if err != nil {
			fmt.Println(err)
		}
	}

	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.PairSuccess:
			fmt.Println("Pair success", client.Store.ID)
			// ws.AddSession(client)
			qrWait <- client.Store.ID.String()
		case *events.Connected:

			fmt.Println("Connected", client.Store.ID)
		case *events.Receipt:
			fmt.Println("Receipt")
			fmt.Println("RECEIPT TYPE", v.Type)
			fmt.Println("RECEIPT IDS", v.MessageIDs)
			fmt.Println("RECEIPT SENDER", v.MessageSender)
			if v.Type == "read" && WaDevice.Webhook != "" {
				body := map[string]any{
					"info":         v,
					"session_name": WaDevice.Session,
				}
				b, _ := json.Marshal(body)

				// fmt.Println(string(b))
				req, err := http.NewRequest("POST", WaDevice.Webhook, bytes.NewBuffer(b))
				if err != nil {
					fmt.Println(err)
				}
				req.Header.Set("Content-Type", "application/json")
				if WaDevice.HeaderKey != "" {
					req.Header.Set("X-Header", WaDevice.HeaderKey)
				}
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
				}
				if resp != nil && resp.Body != nil {
					defer resp.Body.Close()
				}
			}
		case *events.Message:

			var messageBody = v.Message.GetConversation()
			if messageBody == "ping" {
				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
					Conversation: proto.String("pong " + WaDevice.Webhook),
				})

			}
			if WaDevice.Webhook != "" {
				var isDownload = false
				var mediaPath string
				var mimeType, directPath, mmsType string
				var encFileHash, fileHash, mediaKey []byte
				var fileLength int
				var mediaType whatsmeow.MediaType

				// LogJson(v.Message)
				if v.Message.GetImageMessage() != nil {
					img := v.Message.GetImageMessage()
					mediaType = whatsmeow.MediaImage
					directPath = img.GetDirectPath()
					encFileHash = img.GetFileEncSHA256()
					fileHash = img.GetFileSHA256()
					mediaKey = img.GetMediaKey()
					fileLength = int(img.GetFileLength())
					isDownload = true
					mimeType = img.GetMimetype()
				}

				if v.Message.GetVideoMessage() != nil {
					video := v.Message.GetVideoMessage()
					mediaType = whatsmeow.MediaVideo
					directPath = video.GetDirectPath()
					encFileHash = video.GetFileEncSHA256()
					fileHash = video.GetFileSHA256()
					mediaKey = video.GetMediaKey()
					fileLength = int(video.GetFileLength())
					isDownload = true
					mimeType = video.GetMimetype()
				}
				if v.Message.GetAudioMessage() != nil {
					audio := v.Message.GetAudioMessage()
					mediaType = whatsmeow.MediaAudio
					directPath = audio.GetDirectPath()
					encFileHash = audio.GetFileEncSHA256()
					fileHash = audio.GetFileSHA256()
					mediaKey = audio.GetMediaKey()
					fileLength = int(audio.GetFileLength())
					isDownload = true
					mimeType = audio.GetMimetype()
				}
				if v.Message.GetDocumentMessage() != nil {
					doc := v.Message.GetDocumentMessage()
					mediaType = whatsmeow.MediaDocument
					directPath = doc.GetDirectPath()
					encFileHash = doc.GetFileEncSHA256()
					fileHash = doc.GetFileSHA256()
					mediaKey = doc.GetMediaKey()
					fileLength = int(doc.GetFileLength())
					isDownload = true
					mimeType = doc.GetMimetype()
				}

				if isDownload {
					mediaPath2, err := utils.DownloadMedia(client, mimeType, directPath, encFileHash, fileHash, mediaKey, fileLength, mediaType, mmsType)
					if err == nil {
						mediaPath = mediaPath2
					} else {
						fmt.Println("ERROR", err)
					}
				} else {
					mimeType = ""
				}

				body := map[string]any{
					"info":         v.Info,
					"message":      v.Message,
					"sender":       v.Info.Chat.User,
					"jid":          client.Store.ID.String(),
					"session_id":   v.Info.Chat.String(),
					"session_name": WaDevice.Session,
				}
				if mediaPath != "" {
					body["media_path"] = mediaPath
					body["mime_type"] = mimeType

				}

				// utils.LogJson(body)
				b, _ := json.Marshal(body)

				// fmt.Println(string(b))
				req, err := http.NewRequest("POST", WaDevice.Webhook, bytes.NewBuffer(b))
				if err != nil {
					fmt.Println(err)
				}
				req.Header.Set("Content-Type", "application/json")
				if WaDevice.HeaderKey != "" {
					req.Header.Set("X-Header", WaDevice.HeaderKey)
				}
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
				}
				if resp != nil && resp.Body != nil {
					defer resp.Body.Close()
				}

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
