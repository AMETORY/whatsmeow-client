package worker

import (
	"log"
	"os"
	"time"

	"github.com/AMETORY/whatsmeow-client/service"
)

func GetSystemSignal() {
	log.Println("SYSTEM WORKER STARTED")
	dataSub := service.REDIS.Subscribe("SYSTEM")
	for {
		msg, err := dataSub.ReceiveMessage()
		if err != nil {
			log.Println(err)
		}
		if msg.Payload == "RESET" {
			log.Println("SYSTEM WILL RESET IN 5 SECONDS")
			time.Sleep(5 * time.Second)
			os.Exit(0)
		}
	}
}
