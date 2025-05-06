package routes

import (
	"github.com/AMETORY/whatsmeow-client/api/handlers"
	"github.com/AMETORY/whatsmeow-client/objects"

	"github.com/gin-gonic/gin"
)

func NewWaRoutes(r *gin.RouterGroup, sessions *objects.WaSession) {
	handlers := handlers.NewWaHandler(sessions)
	r.POST("/send-message", handlers.SendMessageHandler)
	r.POST("/create-qr", handlers.CreateQRHandler)
	r.GET("/get-qr/:id", handlers.GetQRCodeHandler)
	r.GET("/connected/:id", handlers.CheckConnectedHandler)
	r.DELETE("/disconnect/:id", handlers.DisconnectedHandler)
	r.GET("/check-number/:id/:phone", handlers.IsOnWhatsappHandler)
	r.GET("/devices", handlers.GetDevicesHandler)
	r.PUT("/update-webhook/:id", handlers.UpdateWebhookHandler)
	r.PUT("/message/:id/mark-read", handlers.MarkReadHandler)
	r.GET("/contacts", handlers.GetContactHandler)
	r.GET("/groups/:id", handlers.GetGroupsHandler)
	r.GET("/get-qr-image/:id", handlers.GetQRImageHandler)
	r.GET("/get-group-info/:id/:groupId", handlers.GetGroupInfoHandler)
	r.DELETE("/device-delete/:id", handlers.DeleteDeviceHandler)

}
