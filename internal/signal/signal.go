package signal

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	WS    *websocket.Conn
	Role  string
	Name  string
	User  string
	DeviceID string
	AuthToken string
	ID string
}

type Hub struct {
	mu sync.RWMutex
	clients map[string]*Client
	devices map[string]string
}

func NewHub() *Hub { return &Hub{clients: map[string]*Client{}, devices: map[string]string{}} }

func (h *Hub) Shutdown() {
	for _, c := range h.clients {
		_ = c.WS.Close()
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade websocket: %v", err)
		return
	}
	defer ws.Close()
	clientID := randomClientID()
	client := &Client{WS: ws, ID: clientID, Role: ""}
	h.mu.Lock()
	h.clients[clientID] = client
	h.mu.Unlock()
	_ = ws.SetReadDeadline(time.Now().Add(20 * time.Second))
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		var payload map[string]any
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}
		if payload["type"] == "register" {
			client.Role = payload["role"].(string)
			client.Name = "Guest"
			if name, ok := payload["name"].(string); ok && name != "" { client.Name = name }
			if deviceID, ok := payload["deviceId"].(string); ok && deviceID != "" { client.DeviceID = deviceID }
			_ = ws.WriteJSON(map[string]any{"type": "registered", "id": client.ID})
		}
	}
}

func randomClientID() string { return "client-" + time.Now().Format("20060102150405") }
