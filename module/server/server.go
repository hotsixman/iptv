package server

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
	Engine       *gin.Engine
	Upgrader     *websocket.Upgrader
	Clients      map[*websocket.Conn]bool
	ClientsMutex *sync.Mutex
}

func NewServer() *Server {
	server := &Server{
		Engine: gin.Default(),
		Upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := os.Getenv("ORIGIN")
				if origin == "" {
					return true
				} else {
					return r.Header.Get("Origin") == origin
				}
			},
		},
		Clients:      make(map[*websocket.Conn]bool),
		ClientsMutex: &sync.Mutex{},
	}

	server.Engine.GET("/ws", func(ctx *gin.Context) { handleWS(ctx, server) })
	server.Engine.StaticFile("/", "./index.html")

	return server
}

func handleWS(ctx *gin.Context, server *Server) {
	apiKey := ctx.GetHeader("x-api-key")
	if apiKey != os.Getenv("API_KEY") {
		ctx.AbortWithStatus(401)
		return
	}

	conn, err := server.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.AbortWithStatus(500)
		log.Println(ctx)
		return
	}
	server.ClientsMutex.Lock()
	server.Clients[conn] = true
	server.ClientsMutex.Unlock()
	defer func() {
		server.ClientsMutex.Lock()
		delete(server.Clients, conn)
		server.ClientsMutex.Unlock()
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func BroadcastStream(server *Server, ch chan []byte) {
	for data := range ch {
		server.ClientsMutex.Lock()
		for client := range server.Clients {
			err := client.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				client.Close()
				delete(server.Clients, client)
			}
		}
		server.ClientsMutex.Unlock()
	}
}
