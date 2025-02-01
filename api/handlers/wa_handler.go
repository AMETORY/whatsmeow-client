package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	mdl "github.com/AMETORY/whatsmeow-client/model"
	"github.com/AMETORY/whatsmeow-client/objects"
	"github.com/AMETORY/whatsmeow-client/service"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"github.com/gin-gonic/gin"
	"github.com/mdp/qrterminal"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
)

type WaHandler struct {
	sessions *objects.WaSession
}

func NewWaHandler(sessions *objects.WaSession) *WaHandler {
	return &WaHandler{sessions: sessions}
}

func (wh *WaHandler) GetQRCodeHandler(c *gin.Context) {
	id := c.Param("id")
	res, err := service.REDIS.Get("WA-" + id).Result()
	if err != nil {
		c.JSON(500, gin.H{"message": "failed", "response": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "ok", "response": res})
}
func (wh *WaHandler) DeleteDeviceHandler(c *gin.Context) {
	id := c.Param("id")
	jid, err := types.ParseJID(id)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 1", "response": err.Error()})
		return
	}

	deviceStore, err := wh.sessions.Container.GetDevice(jid)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 2", "response": err.Error()})
		return
	}

	var client *whatsmeow.Client
	for _, v := range wh.sessions.Clients {

		if v.Store.ID.String() == jid.String() {
			client = v
		}
	}

	if client == nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": "client not found"})
		return

	}
	client.Disconnect()
	err = wh.sessions.Container.DeleteDevice(deviceStore)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

func (wh *WaHandler) GetQRImageHandler(c *gin.Context) {
	id := c.Param("id")
	res, err := service.REDIS.Get("WA-" + id).Result()
	if err != nil {
		c.JSON(500, gin.H{"message": "failed", "response": err.Error()})
		return
	}

	qrCode, err := qrcode.Encode(res, qrcode.Medium, 256)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate QR code"})
		return
	}
	img := "data:image/png;base64," + base64.StdEncoding.EncodeToString(qrCode)

	response := []byte(
		fmt.Sprintf("<p><img src='%s' /></p><script>setTimeout(function(){ location.reload(); }, 60000);</script>", img),
	)
	c.Data(200, http.DetectContentType(response), response)
}

func (wh *WaHandler) CreateQRHandler(c *gin.Context) {
	var input mdl.WaDevice
	err := c.ShouldBindBodyWithJSON(&input)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 0", "response": err.Error()})
		return
	}
	qrWait := make(chan string)
	// qrCode := make(chan string)
	deviceStore := wh.sessions.Container.NewDevice()
	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(wh.sessions.GetEventHandler(client, qrWait))
	qrChan, _ := client.GetQRChannel(wh.sessions.Ctx)
	err = client.Connect()
	if err != nil {
		panic(err)
	}

	for evt := range qrChan {
		if evt.Event == "code" {
			service.REDIS.Set("WA-"+input.Session, evt.Code, 0)
			// Render the QR code here
			qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal:
			// fmt.Println("QR code:", evt.Code)
		} else {
			fmt.Println("Login event:", evt.Event)
		}
	}
	response := <-qrWait
	client.Connect()
	wh.sessions.AddSession(client)
	service.REDIS.Del("WA-" + input.Session)
	input.JID = client.Store.ID.String()
	wh.sessions.DB.Create(&input)
	c.JSON(200, gin.H{"message": "ok", "response": response})
}
func (wh *WaHandler) SendMessageHandler(c *gin.Context) {

	var input objects.WaMessage
	err := c.ShouldBindBodyWithJSON(&input)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 0", "response": err.Error()})
		return
	}

	jid, err := types.ParseJID(input.JID)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 1", "response": err.Error()})
		return
	}

	fmt.Println("GET JID", jid)
	// deviceStore, err := wh.sessions.Container.GetDevice(jid)
	// if err != nil {
	// 	c.JSON(500, gin.H{"message": "failed 2", "response": err.Error()})
	// 	return
	// }
	var client *whatsmeow.Client
	for _, v := range wh.sessions.Clients {

		if v.Store.ID.String() == jid.String() {
			client = v
		}
	}

	// fmt.Println("CLIENT", client)
	// clientLog := waLog.Stdout("Client", "INFO", true)
	// client := whatsmeow.NewClient(deviceStore, clientLog)

	recipient, err := types.ParseJID(input.To + "@s.whatsapp.net")
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 4", "response": err.Error()})
		return
	}
	var dataMessage *waE2E.Message = &waE2E.Message{}
	if input.FileType != "" && input.FileUrl != "" {
		resp, err := http.Get(input.FileUrl)
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to fetch file", "response": err.Error()})
			return
		}
		defer resp.Body.Close()

		fileBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to read file", "response": err.Error()})
			return
		}

		mimeType := http.DetectContentType(fileBytes)

		var fileType whatsmeow.MediaType
		switch input.FileType {
		case "image":
			fileType = whatsmeow.MediaImage
		case "video":
			fileType = whatsmeow.MediaVideo
		case "audio":
			fileType = whatsmeow.MediaAudio
		case "document":
			fileType = whatsmeow.MediaDocument
		}

		respUpload, err := client.Upload(wh.sessions.Ctx, fileBytes, fileType)
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to upload file", "response": err.Error()})
			return
		}

		switch fileType {
		case whatsmeow.MediaImage:
			dataMessage.ImageMessage = &waE2E.ImageMessage{
				Caption:       proto.String(input.Text),
				Mimetype:      proto.String(mimeType),
				URL:           &respUpload.URL,
				DirectPath:    &respUpload.DirectPath,
				MediaKey:      respUpload.MediaKey,
				FileEncSHA256: respUpload.FileEncSHA256,
				FileSHA256:    respUpload.FileSHA256,
				FileLength:    &respUpload.FileLength,
			}
		case whatsmeow.MediaVideo:
			dataMessage.VideoMessage = &waE2E.VideoMessage{
				Caption:       proto.String(input.Text),
				Mimetype:      proto.String(mimeType),
				URL:           &respUpload.URL,
				DirectPath:    &respUpload.DirectPath,
				MediaKey:      respUpload.MediaKey,
				FileEncSHA256: respUpload.FileEncSHA256,
				FileSHA256:    respUpload.FileSHA256,
				FileLength:    &respUpload.FileLength,
			}
		case whatsmeow.MediaAudio:
			dataMessage.AudioMessage = &waE2E.AudioMessage{
				Mimetype:      proto.String(mimeType),
				URL:           &respUpload.URL,
				DirectPath:    &respUpload.DirectPath,
				MediaKey:      respUpload.MediaKey,
				FileEncSHA256: respUpload.FileEncSHA256,
				FileSHA256:    respUpload.FileSHA256,
				FileLength:    &respUpload.FileLength,
			}
		case whatsmeow.MediaDocument:
			dataMessage.DocumentMessage = &waE2E.DocumentMessage{
				Caption:       proto.String(input.Text),
				Mimetype:      proto.String(mimeType),
				URL:           &respUpload.URL,
				DirectPath:    &respUpload.DirectPath,
				MediaKey:      respUpload.MediaKey,
				FileEncSHA256: respUpload.FileEncSHA256,
				FileSHA256:    respUpload.FileSHA256,
				FileLength:    &respUpload.FileLength,
			}
		default:
			dataMessage.Conversation = proto.String(input.Text)
		}

	}
	resp, err := client.SendMessage(wh.sessions.Ctx, recipient, dataMessage)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 5", "response": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "ok", "jid": resp})
}
