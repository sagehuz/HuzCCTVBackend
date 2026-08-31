package signal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestHubRelayAndDeviceList exercises the exact signaling path the Phone
// Remote page (/phone.html) relies on:
//   - a device registers with kind/model and appears in the viewer's device-list
//   - viewer -> device control messages are relayed verbatim (with "from")
//   - device -> viewer capability replies are relayed back
func TestHubRelayAndDeviceList(t *testing.T) {
	hub := NewHub(func(string) bool { return true })
	srv := httptest.NewServer(http.HandlerFunc(hub.HandleWS))
	defer srv.Close()
	defer hub.Shutdown()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/signal"

	// Device: no auth cookie required.
	devConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("device dial: %v", err)
	}
	defer devConn.Close()

	// Viewer: must present a valid session cookie (validator always accepts here).
	viewConn, _, err := websocket.DefaultDialer.Dial(
		wsURL,
		http.Header{"Cookie": []string{"huz_session=testtoken"}},
	)
	if err != nil {
		t.Fatalf("viewer dial: %v", err)
	}
	defer viewConn.Close()

	mustWrite(t, devConn, map[string]any{
		"type": "register", "role": "device", "deviceId": "mock-phone",
		"name": "Test Phone", "kind": "phone", "model": "Pixel 4a",
	})
	// Drain the device's own "registered" ack.
	_ = devConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var regAck map[string]any
	if err := devConn.ReadJSON(&regAck); err != nil {
		t.Fatalf("device registered ack: %v", err)
	}
	if regAck["type"] != "registered" {
		t.Fatalf("expected registered ack, got %v", regAck)
	}
	mustWrite(t, viewConn, map[string]any{"type": "register", "role": "viewer"})

	// Viewer should receive device-list containing the phone with kind/model.
	var deviceID string
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && deviceID == "" {
		_ = viewConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var msg map[string]any
		if err := viewConn.ReadJSON(&msg); err != nil {
			t.Fatalf("viewer read: %v", err)
		}
		if msg["type"] != "device-list" {
			continue
		}
		devices, _ := msg["devices"].([]any)
		if len(devices) != 1 {
			t.Fatalf("expected 1 device in list, got %d", len(devices))
		}
		d, _ := devices[0].(map[string]any)
		if d["kind"] != "phone" || d["model"] != "Pixel 4a" {
			t.Fatalf("device kind/model not forwarded: %v", d)
		}
		deviceID, _ = d["id"].(string)
	}
	if deviceID == "" {
		t.Fatal("viewer never learned the device id")
	}

	// Viewer -> device control relay.
	mustWrite(t, viewConn, map[string]any{"type": "control", "targetId": deviceID, "action": "screen-on"})
	_ = devConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ctrl map[string]any
	if err := devConn.ReadJSON(&ctrl); err != nil {
		t.Fatalf("device read control: %v", err)
	}
	if ctrl["type"] != "control" || ctrl["action"] != "screen-on" {
		t.Fatalf("control not relayed verbatim: %v", ctrl)
	}
	fromID, _ := ctrl["from"].(string)
	if fromID == "" {
		t.Fatal("control missing 'from'")
	}

	// Device -> viewer capabilities reply.
	mustWrite(t, devConn, map[string]any{
		"type": "capabilities", "targetId": fromID,
		"data": map[string]any{
			"model":      "Pixel 4a",
			"screenSize": map[string]any{"width": 1080, "height": 2400},
		},
	})
	_ = viewConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var capMsg map[string]any
	if err := viewConn.ReadJSON(&capMsg); err != nil {
		t.Fatalf("viewer read capabilities: %v", err)
	}
	if capMsg["type"] != "capabilities" {
		t.Fatalf("expected capabilities reply, got %v", capMsg)
	}
	data, _ := capMsg["data"].(map[string]any)
	if data["model"] != "Pixel 4a" {
		t.Fatalf("capabilities data not relayed verbatim: %v", data)
	}
}

func mustWrite(t *testing.T, conn *websocket.Conn, v any) {
	t.Helper()
	if err := conn.WriteJSON(v); err != nil {
		t.Fatalf("write: %v", err)
	}
}
