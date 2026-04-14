package main

import (
	"homecam/module/ffmpeg"
	"homecam/module/types"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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
	go captureFrame()
	go broadcastStream()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r)
	})

	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	port := ":3000"
	log.Printf("Server started on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	clientsMux.Lock()
	clients[conn] = true
	clientsMux.Unlock()

	log.Printf("New client connected. Total clients: %d", len(clients))

	defer func() {
		clientsMux.Lock()
		delete(clients, conn)
		clientsMux.Unlock()
		conn.Close()
		log.Println("Client disconnected")
	}()

	// Keep connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func broadcastStream() {
	for data := range dataChan {
		clientsMux.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		clientsMux.Unlock()
	}
}

func captureFrame() {
	device := types.Device{Name: os.Getenv("DEVICE")}

	width, err := strconv.Atoi(os.Getenv("WIDTH"))
	height, err := strconv.Atoi(os.Getenv("HEIGHT"))
	fps, err := strconv.ParseFloat(os.Getenv("FPS"), 64)
	format := types.Format{
		Codec:  os.Getenv("CODEC"),
		Width:  width,
		Height: height,
		Fps:    fps,
	}

	log.Printf("Starting H.264 capture: %s (%dx%d)", device.Name, format.Width, format.Height)

	cmd := ffmpeg.MakeExecH264(device, format)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 1024*32) // 32KB chunks
	for {
		n, err := stdout.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println("Read error:", err)
			}
			break
		}

		if n > 0 {
			// Send copy of data to channel
			data := make([]byte, n)
			copy(data, buf[:n])

			select {
			case dataChan <- data:
			default:
				// Skip if channel is full to prevent blocking
			}
		}
	}
}
