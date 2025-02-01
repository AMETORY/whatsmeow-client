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
	r.GET("/get-qr-image/:id", handlers.GetQRImageHandler)
	r.DELETE("/device-delete/:id", handlers.DeleteDeviceHandler)

}
