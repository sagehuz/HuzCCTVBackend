package signal

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	WS        *websocket.Conn
	Role      string
	Name      string
	User      string
	DeviceID  string
	AuthToken string
	ID        string
	alive     bool
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
	devices map[string]string
	isValidToken func(string) bool
}

func NewHub(validate func(string) bool) *Hub {
	return &Hub{clients: map[string]*Client{}, devices: map[string]string{}, isValidToken: validate}
}

func (h *Hub) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, c := range h.clients {
		if c != nil && c.WS != nil {
			_ = c.WS.Close()
		}
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade websocket: %v", err)
		return
	}
	ws.SetReadLimit(1 << 20)
	clientID := newClientID()
	client := &Client{WS: ws, ID: clientID, Name: "Thiết bị " + clientID[:6], alive: true}
	if cookie, err := r.Cookie("huz_session"); err == nil {
		client.AuthToken = strings.TrimSpace(cookie.Value)
	}
	h.addClient(clientID, client)
	defer h.removeClient(clientID)
	defer ws.Close()

	ws.SetPingHandler(func(app string) error {
		client.alive = true
		return ws.WriteControl(websocket.PongMessage, []byte(app), time.Now().Add(time.Second))
	})
	ws.SetPongHandler(func(string) error {
		client.alive = true
		return nil
	})
	go h.pingLoop(client)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		if len(msg) == 0 {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}
		if payload["type"] == nil {
			continue
		}
		typeName, _ := payload["type"].(string)
		switch typeName {
		case "register":
			h.handleRegister(client, payload, r)
		case "watch", "offer", "answer", "ice-candidate", "control", "capabilities", "device-status", "snapshot":
			h.handleRelay(client, payload)
		}
	}
}

func (h *Hub) addClient(id string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[id] = c
}

func (h *Hub) removeClient(id string) {
	h.mu.Lock()
	if c, ok := h.clients[id]; ok {
		if c != nil && c.Role == "device" && c.DeviceID != "" {
			if cur, ok := h.devices[c.DeviceID]; ok && cur == id {
				delete(h.devices, c.DeviceID)
			}
		}
		delete(h.clients, id)
	}
	h.mu.Unlock()
	h.broadcastDeviceList()
}

func (h *Hub) handleRegister(client *Client, payload map[string]any, r *http.Request) {
	role, _ := payload["role"].(string)
	if role != "device" && role != "viewer" {
		return
	}
	if role == "viewer" {
		cookie, err := r.Cookie("huz_session")
		if err != nil || !h.tokenValid(strings.TrimSpace(cookie.Value)) {
			h.sendError(client, "Bạn cần đăng nhập để xem camera")
			h.forceClose(client, 4001, "Unauthorized")
			return
		}
		client.AuthToken = strings.TrimSpace(cookie.Value)
	}
	client.Role = role
	client.Name = "Thiết bị " + client.ID[:6]
	if name, ok := payload["name"].(string); ok {
		name = strings.TrimSpace(name)
		if len(name) > 64 {
			name = name[:64]
		}
		if name != "" {
			client.Name = name
		}
	}
	if role == "device" {
		deviceID, _ := payload["deviceId"].(string)
		deviceID = strings.TrimSpace(deviceID)
		if deviceID == "" {
			deviceID = "d_" + client.ID
		}
		if len(deviceID) > 128 {
			deviceID = deviceID[:128]
		}
		client.DeviceID = deviceID
		h.mu.Lock()
		if oldID, ok := h.devices[deviceID]; ok && oldID != client.ID {
			if oldClient, exists := h.clients[oldID]; exists && oldClient != nil {
				h.mu.Unlock()
				h.sendError(oldClient, "Thiết bị đã kết nối lại ở phiên mới, đóng kết nối cũ")
				h.forceClose(oldClient, 4002, "replaced")
				h.mu.Lock()
			}
			delete(h.devices, deviceID)
		}
		h.devices[deviceID] = client.ID
		h.mu.Unlock()
	}
	_ = client.WS.WriteJSON(map[string]any{"type": "registered", "id": client.ID})
	h.broadcastDeviceList()
}

func (h *Hub) handleRelay(client *Client, payload map[string]any) {
	if client == nil || client.WS == nil {
		return
	}
	targetID, _ := payload["targetId"].(string)
	targetID = strings.TrimSpace(targetID)
	if targetID == "" {
		return
	}
	payload = cloneMessage(payload)
	payload["from"] = client.ID
	h.mu.RLock()
	target, ok := h.clients[targetID]
	h.mu.RUnlock()
	if !ok || target == nil || target.WS == nil {
		h.sendError(client, "Thiết bị đích không còn kết nối")
		return
	}
	if err := target.WS.WriteJSON(payload); err != nil {
		h.sendError(client, "Thiết bị đích không còn kết nối")
	}
}

func (h *Hub) pingLoop(client *Client) {
	for {
		time.Sleep(10 * time.Second)
		if client == nil || client.WS == nil {
			return
		}
		if !client.alive {
			_ = client.WS.Close()
			return
		}
		client.alive = false
		if err := client.WS.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
			_ = client.WS.Close()
			return
		}
	}
}

func (h *Hub) sendError(client *Client, msg string) {
	if client == nil || client.WS == nil {
		return
	}
	_ = client.WS.WriteJSON(map[string]any{"type": "error", "message": msg})
}

func (h *Hub) forceClose(client *Client, code int, reason string) {
	if client == nil || client.WS == nil {
		return
	}
	go func() {
		time.Sleep(200 * time.Millisecond)
		_ = client.WS.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(time.Second))
		_ = client.WS.Close()
	}()
}

func (h *Hub) broadcastDeviceList() {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for _, c := range h.clients {
		if c != nil && c.Role == "device" {
			clients = append(clients, c)
		}
	}
	viewerIDs := make([]string, 0)
	for id, c := range h.clients {
		if c != nil && c.Role == "viewer" {
			viewerIDs = append(viewerIDs, id)
		}
	}
	payload := map[string]any{"type": "device-list", "devices": []map[string]any{}}
	list := make([]map[string]any, 0, len(clients))
	for _, c := range clients {
		list = append(list, map[string]any{"id": c.ID, "name": c.Name, "deviceId": c.DeviceID})
	}
	payload["devices"] = list
	for _, id := range viewerIDs {
		if c, ok := h.clients[id]; ok && c != nil && c.WS != nil {
			_ = c.WS.WriteJSON(payload)
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) tokenValid(raw string) bool {
	decoded, err := url.QueryUnescape(strings.TrimSpace(raw))
	if err != nil {
		decoded = strings.TrimSpace(raw)
	}
	if h.isValidToken == nil {
		return false
	}
	return h.isValidToken(decoded)
}

func cloneMessage(in map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func newClientID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

var errUnauthorized = errors.New("unauthorized")
