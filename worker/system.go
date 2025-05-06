package worker

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AMETORY/whatsmeow-client/service"
)

func GetSystemSignal() {
	fmt.Println("SYSTEM WORKER STARTED")
	dataSub := service.REDIS.Subscribe("SYSTEM")
	for {
		msg, err := dataSub.ReceiveMessage()
		if err != nil {
			log.Println(err)
		}
		if msg.Payload == "RESET" {
			fmt.Println("SYSTEM WILL RESET IN 5 SECONDS")
			time.Sleep(5 * time.Second)
			os.Exit(0)
		}
	}
}
