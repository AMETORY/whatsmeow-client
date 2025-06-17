package handlers

import (
	bgContext "context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	mdl "github.com/AMETORY/whatsmeow-client/model"
	"github.com/AMETORY/whatsmeow-client/objects"
	"github.com/AMETORY/whatsmeow-client/service"
	"github.com/AMETORY/whatsmeow-client/utils"
	"github.com/gabriel-vasile/mimetype"

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

func (wh *WaHandler) GetDevicesHandler(c *gin.Context) {
	var devices []mdl.WaDevice
	wh.sessions.DB.Find(&devices)
	c.JSON(200, gin.H{"message": "ok", "data": devices})
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
func (wh *WaHandler) getClient(id string) *whatsmeow.Client {
	jid, err := types.ParseJID(id)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for _, v := range wh.sessions.Clients {
		if v == nil {
			continue
		}
		if v.Store == nil {
			continue
		}
		if v.Store.ID == nil {
			continue
		}
		if v.Store.ID.String() == jid.String() {
			return v
		}
	}
	// deviceStore, err := wh.sessions.Container.GetDevice(jid)
	// if err == nil {
	// 	clientLog := waLog.Stdout("Client ["+deviceStore.ID.String()+"]", "INFO", true)
	// 	client := whatsmeow.NewClient(deviceStore, clientLog)
	// 	client.Connect()
	// 	// client.AddEventHandler(wh.sessions.GetEventHandler(client, nil))
	// 	return client
	// }

	return nil
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

	var client *whatsmeow.Client = wh.getClient(id)
	if client == nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": "client not found"})
		return

	}
	err = client.Logout()
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 4", "response": err.Error()})
		return
	}
	err = wh.sessions.Container.DeleteDevice(deviceStore)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": err.Error()})
		return
	}

	var WaDevice mdl.WaDevice
	err = wh.sessions.DB.Where("j_id = ?", id).Delete(&WaDevice).Error
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": err.Error()})
		return
	}
	service.REDIS.Publish("SYSTEM", "RESET")

	c.JSON(200, gin.H{"message": "ok"})
}

func (wh *WaHandler) GetJIDfromSessionHandler(c *gin.Context) {
	session := c.Param("session")
	var data mdl.WaDevice
	err := wh.sessions.DB.Where("session = ?", session).First(&data).Error
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 1", "response": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "ok", "jid": data.JID})
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

func (wh *WaHandler) GetGroupInfoHandler(c *gin.Context) {
	id := c.Param("id")

	groupId := c.Param("groupId")
	jgroupId, err := types.ParseJID(groupId)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 1", "response": err.Error()})
		return
	}

	var client *whatsmeow.Client = wh.getClient(id)
	if client == nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": "client not found"})
		return

	}
	info, err := client.GetGroupInfo(jgroupId)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "ok", "data": info})
}

func (wh *WaHandler) GetGroupsHandler(c *gin.Context) {
	id := c.Param("id")

	var client *whatsmeow.Client = wh.getClient(id)
	if client == nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": "client not found"})
		return

	}

	groups, err := client.GetJoinedGroups()
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "ok", "data": groups})
}
func (wh *WaHandler) GetContactHandler(c *gin.Context) {
	var contacts []mdl.WhatsmeowContact
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(500, gin.H{"message": "failed", "error": err.Error()})
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(500, gin.H{"message": "failed", "error": err.Error()})
		return
	}
	offset := (page - 1) * pageSize
	var count int64
	err = wh.sessions.DB.Model(&mdl.WhatsmeowContact{}).Count(&count).Error
	if err != nil {
		c.JSON(500, gin.H{"message": "failed", "error": err.Error()})
		return
	}
	JID := c.DefaultQuery("jid", "")
	if JID == "" {
		c.JSON(500, gin.H{"message": "failed", "error": "JID is required"})
		return
	}

	hasNext := count > int64((page * pageSize))
	hasPrevious := page > 1
	search := c.DefaultQuery("search", "")
	err = wh.sessions.DB.Where("our_jid = ?", JID).Where("( their_jid LIKE ? OR full_name LIKE ? OR push_name LIKE ? OR business_name LIKE ?)",
		"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%").Offset(offset).Limit(pageSize).Find(&contacts).Error
	if err != nil {
		c.JSON(500, gin.H{"message": "failed", "response": err.Error()})
		return
	}

	for i, v := range contacts {
		userJid, _ := types.ParseJID(v.TheirJid)
		v.PhoneNumber = userJid.User

		contacts[i] = v
	}

	c.JSON(200, gin.H{"message": "ok", "data": gin.H{"items": contacts, "total": count, "page": page, "limit": pageSize, "has_next": hasNext, "has_previous": hasPrevious}})
}
func (wh *WaHandler) MarkReadHandler(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		MsgIDs []string `json:"msg_ids"`
		ChatID string   `json:"chat_id"`
	}
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 0", "response": err.Error()})
		return
	}
	var data mdl.WaDevice
	err = wh.sessions.DB.Where("j_id = ? OR session = ?", id, id).First(&data).Error
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 1", "response": err.Error()})
		return
	}

	var client *whatsmeow.Client = wh.getClient(id)
	if client == nil {
		c.JSON(500, gin.H{"message": "failed 2", "response": "client not found"})
		return

	}
	ids := []types.MessageID{}
	for _, v := range input.MsgIDs {
		ids = append(ids, types.MessageID(v))
	}
	chatID, _ := types.ParseJID(input.ChatID + "@s.whatsapp.net")
	senderID, _ := types.ParseJID(data.JID)
	sender := senderID
	if strings.ContainsRune(data.JID, ':') {
		parts := strings.Split(data.JID, ":")
		newSender, _ := types.ParseJID(parts[0] + "@s.whatsapp.net")
		sender = newSender
	}
	fmt.Println(ids, time.Now(), sender, chatID, types.ReceiptTypeRead)
	err = client.MarkRead(ids, time.Now(), chatID, sender, types.ReceiptTypeRead)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 1", "response": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "ok"})
}
func (wh *WaHandler) UpdateWebhookHandler(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Webhook   string `json:"webhook"`
		HeaderKey string `json:"header_key"`
	}
	var data mdl.WaDevice
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 0", "response": err.Error()})
		return
	}

	err = wh.sessions.DB.Where("j_id = ? OR session = ?", id, id).First(&data).Error
	if err != nil {
		fmt.Println(err)
	}

	data.Webhook = input.Webhook
	data.HeaderKey = input.HeaderKey
	err = wh.sessions.DB.Where("j_id = ? OR session = ?", id, id).Save(&data).Error
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(200, gin.H{"message": "ok"})

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
	fmt.Println("client", client)
	if client != nil {
		client.AddEventHandler(wh.sessions.GetEventHandler(client, qrWait))
	}
	ctx, cancel := bgContext.WithTimeout(wh.sessions.Ctx, 30*time.Second)
	defer cancel()
	qrChan, _ := client.GetQRChannel(ctx)
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
	client.AddEventHandler(wh.sessions.GetEventHandler(client, qrWait))
	fmt.Println("CLIENT PAIRED", client.Store.ID.String(), client.IsConnected())
	wh.sessions.AddSession(client)
	service.REDIS.Del("WA-" + input.Session)
	input.JID = client.Store.ID.String()
	err = wh.sessions.DB.Create(&input).Error
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 1", "response": err.Error()})
		return
	}
	service.REDIS.Publish("SYSTEM", "RESET")
	c.JSON(200, gin.H{"message": "ok", "response": response, "data": input})
}
func (wh *WaHandler) SendMessageHandler(c *gin.Context) {

	var input objects.WaMessage
	err := c.ShouldBindBodyWithJSON(&input)
	if err != nil {
		fmt.Println("ERROR #1", err.Error())
		c.JSON(500, gin.H{"message": "failed 0", "response": err.Error()})
		return
	}

	utils.LogJson(input)

	var client *whatsmeow.Client = wh.getClient(input.JID)
	if client == nil {
		fmt.Println("ERROR #2", "NO CLIENT")
		c.JSON(500, gin.H{"message": "failed 3", "response": "client not found"})
		return

	}

	// fmt.Println("CLIENT", client)
	// clientLog := waLog.Stdout("Client", "INFO", true)
	// client := whatsmeow.NewClient(deviceStore, clientLog)
	receiver := input.To + "@s.whatsapp.net"
	if input.IsGroup {
		receiver = input.To
	}
	recipient, err := types.ParseJID(receiver)
	if err != nil {
		fmt.Println("ERROR #3", err.Error())
		c.JSON(500, gin.H{"message": "failed 4", "response": err.Error()})
		return
	}
	var dataMessage *waE2E.Message = &waE2E.Message{
		Conversation: proto.String(input.Text),
	}
	if input.FileType != "" && input.FileUrl != "" {
		resp, err := http.Get(input.FileUrl)
		if err != nil {
			fmt.Println("ERROR #4", err.Error())
			c.JSON(500, gin.H{"message": "failed to fetch file", "response": err.Error()})
			return
		}
		defer resp.Body.Close()

		fileBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("ERROR #4", err.Error())
			c.JSON(500, gin.H{"message": "failed to read file", "response": err.Error()})
			return
		}

		mtype := mimetype.Detect(fileBytes)

		mimeType := mtype.String()

		fmt.Println("MIME TYPE", mimeType)

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
		default:

		}

		respUpload, err := client.Upload(wh.sessions.Ctx, fileBytes, fileType)
		if err != nil {
			fmt.Println("ERROR #4", err.Error())
			c.JSON(500, gin.H{"message": "failed to upload file", "response": err.Error()})
			return
		}

		// Extract the file name from the URL

		switch fileType {
		case whatsmeow.MediaImage:
			dataMessage.Conversation = nil
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
			dataMessage.Conversation = nil
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
			dataMessage.Conversation = nil
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
			fileName := input.FileUrl[strings.LastIndex(input.FileUrl, "/")+1:]
			dataMessage.Conversation = nil
			dataMessage.DocumentMessage = &waE2E.DocumentMessage{
				Caption:       proto.String(input.Text),
				Mimetype:      proto.String(mimeType),
				URL:           &respUpload.URL,
				DirectPath:    &respUpload.DirectPath,
				MediaKey:      respUpload.MediaKey,
				FileEncSHA256: respUpload.FileEncSHA256,
				FileSHA256:    respUpload.FileSHA256,
				FileLength:    &respUpload.FileLength,
				FileName:      proto.String(fileName),
			}

		}

	}

	// fmt.Println("recipient", recipient)

	// dataMessage.TemplateMessage = &waE2E.TemplateMessage{
	// 	HydratedTemplate: &waE2E.TemplateMessage_HydratedFourRowTemplate{
	// 		TemplateID: proto.String("template-1"),
	// 		Title: &waE2E.TemplateMessage_HydratedFourRowTemplate_HydratedTitleText{
	// 			HydratedTitleText: "Hello",
	// 		},
	// 		HydratedButtons: []*waE2E.HydratedTemplateButton{
	// 			{
	// 				Index: proto.Uint32(1),
	// 				HydratedButton: &waE2E.HydratedTemplateButton_QuickReplyButton{
	// 					QuickReplyButton: &waE2E.HydratedTemplateButton_HydratedQuickReplyButton{
	// 						DisplayText: proto.String("Apa Kabar?"),
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	// LogJson(dataMessage)
	if input.RefID != nil && input.RefFrom != nil && input.RefText != nil {
		dataMessage.Conversation = nil
		dataMessage.ExtendedTextMessage = &waE2E.ExtendedTextMessage{
			Text: proto.String(input.Text),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(*input.RefID),
				Participant: proto.String(*input.RefFrom),
				QuotedMessage: &waE2E.Message{
					Conversation: proto.String(*input.RefText),
				},
			},
		}
	}
	utils.LogJson(dataMessage)
	resp, err := client.SendMessage(wh.sessions.Ctx, recipient, dataMessage)
	if err != nil {
		fmt.Println("ERROR #5", err.Error())
		c.JSON(500, gin.H{"message": "failed 5", "response": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "ok", "data": resp})
}

func (wh *WaHandler) CheckConnectedHandler(c *gin.Context) {
	id := c.Param("id")

	var client *whatsmeow.Client = wh.getClient(id)
	if client == nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": "client not found"})
		return

	}
	c.JSON(200, gin.H{"message": "ok", "is_connected": client.IsConnected()})
}
func (wh *WaHandler) DisconnectedHandler(c *gin.Context) {
	id := c.Param("id")

	var client *whatsmeow.Client = wh.getClient(id)
	if client == nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": "client not found"})
		return

	}

	// client.Disconnect()
	err := client.Logout()
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 4", "response": err.Error()})
		return
	}
	service.REDIS.Publish("SYSTEM", "RESET")
	c.JSON(200, gin.H{"message": "ok"})
}

func (wh *WaHandler) IsOnWhatsappHandler(c *gin.Context) {
	phone := c.Param("phone")
	id := c.Param("id")

	var client *whatsmeow.Client = wh.getClient(id)
	if client == nil {
		c.JSON(500, gin.H{"message": "failed 3", "response": "client not found"})
		return

	}

	isOnWhatsapp, err := client.IsOnWhatsApp([]string{phone})
	if err != nil {
		c.JSON(500, gin.H{"message": "failed 4", "response": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "ok", "is_on_whatsapp": isOnWhatsapp})
}
func LogJson(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
}
