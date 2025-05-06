package main

import (
	"context"
	"fmt"
	"io"

	"github.com/AMETORY/whatsmeow-client/api/routes"
	mdl "github.com/AMETORY/whatsmeow-client/model"
	"github.com/AMETORY/whatsmeow-client/objects"
	"github.com/AMETORY/whatsmeow-client/service"
	"github.com/AMETORY/whatsmeow-client/worker"

	"github.com/gin-gonic/gin"
	// _ "modernc.org/sqlite"

	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow"

	// waE2E "go.mau.fi/whatsmeow/binary/proto"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	ctx := context.Background()

	service.InitRedis()
	db, err := service.InitDB()
	if err != nil {
		panic(err)
	}

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite as we did in this minimal working example
	container, err := sqlstore.New("postgres", "user=postgres dbname=whatsapp sslmode=disable password=balakutak", dbLog)
	// container, err := sqlstore.New("sqlite3", "file:wa.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	sessions := objects.NewWaSession(ctx, container, db)
	deviceStore, err := container.GetAllDevices()
	if err != nil {
		panic(err)
	}
	for _, device := range deviceStore {
		clientLog := waLog.Stdout("Client ["+device.ID.String()+"]", "INFO", true)
		client := whatsmeow.NewClient(device, clientLog)
		client.Connect()
		client.AddEventHandler(sessions.GetEventHandler(client, nil))
		sessions.AddSession(client)
	}

	fmt.Printf("ADD %v DEVICE", len(deviceStore))

	err = db.AutoMigrate(&mdl.WaDevice{})
	if err != nil {
		panic(err)
	}

	// container, err := sqlstore.New("sqlite3", "file:wa.db?_foreign_keys=on", dbLog)
	// if err != nil {
	// 	panic(err)
	// })

	// // If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	// deviceStore, err := container.GetFirstDevice()
	// if err != nil {
	// 	panic(err)
	// }

	// clientLog := waLog.Stdout("Client", "INFO", true)
	// client := whatsmeow.NewClient(deviceStore, clientLog)
	// client.AddEventHandler(GetEventHandler(client))

	// if client.Store.ID == nil {
	// 	// No ID stored, new login
	// 	qrChan, _ := client.GetQRChannel(context.Background())
	// 	err = client.Connect()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	for evt := range qrChan {
	// 		if evt.Event == "code" {
	// 			// Render the QR code here
	// 			qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
	// 			// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal:
	// 			// fmt.Println("QR code:", evt.Code)
	// 		} else {
	// 			fmt.Println("Login event:", evt.Event)
	// 		}
	// 	}
	// } else {
	// 	// Already logged in, just connect
	// 	err = client.Connect()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	r := gin.Default()

	v1 := r.Group("/v1")

	routes.NewWaRoutes(v1, sessions)
	r.Static("/static", "./tmp")

	r.POST("/webhook", func(ctx *gin.Context) {
		req := ctx.Request.Body
		body, err := io.ReadAll(req)
		if err != nil {
			fmt.Println(err)
		}
		strBody := string(body)

		fmt.Println(strBody)
	})

	go func() {
		worker.GetSystemSignal()
	}()

	r.Run(":8088")

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// <-c

	// client.Disconnect()
}
