package utils

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	"go.mau.fi/whatsmeow"
)

func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "video/mp4":
		return ".mp4"
	case "audio/ogg; codecs=opus", "audio/ogg":
		return ".ogg"
	case "audio/mp3":
		return ".mp3"
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "application/vnd.ms-excel":
		return ".xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	case "application/vnd.ms-powerpoint":
		return ".ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return ".pptx"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
	}
}

func DownloadMedia(client *whatsmeow.Client, mimeType, directPath string, encFileHash, fileHash, mediaKey []byte, fileLength int, mediaType whatsmeow.MediaType, mmsType string) (mediaPath string, err error) {
	dataImg, err := client.DownloadMediaWithPath(
		client.BackgroundEventCtx,
		directPath,
		encFileHash,
		fileHash,
		mediaKey,
		fileLength,
		mediaType,
		mmsType,
	)
	if err == nil {
		// Save image to disk
		// SaveImage(dataImg, "image.jpg")
		if _, err := os.Stat("tmp"); os.IsNotExist(err) {
			err = os.Mkdir("tmp", 0755)
			if err != nil {
				log.Println(err)
			}
		}
		filename := RandomString(10) + GetExtensionFromMimeType(mimeType)
		mediaPath = "/static/" + filename
		err = os.WriteFile("tmp/"+filename, dataImg, 0644)
		if err != nil {
			log.Println(err)

		}
	}
	return
}

func LogJson(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(string(data))
}
func SaveJson(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(string(data))
}
