package main

import (
	"homecam/module/ffmpeg"
	"homecam/module/server"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	clients    = make(map[*websocket.Conn]bool)
	clientsMux sync.Mutex
	dataChan   = make(chan []byte, 100) // Channel for H.264 stream data
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	device, format, err := ffmpeg.LoadDeviceAndForamtFromEnv()
	if err != nil {
		log.Fatalln(err)
		return
	}

	app := server.NewServer()
	ch := make(chan []byte)

	go ffmpeg.CaptureFrame(*device, *format, ch, 10)
	go server.BroadcastStream(app, ch)

	app.Engine.Run(os.Getenv("PORT"))
}
