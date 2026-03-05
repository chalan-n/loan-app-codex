// handlers/websocket.go
package handlers

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type wsClient struct {
	conn    *websocket.Conn
	staffID string
}

var (
	wsClients   = make(map[*websocket.Conn]*wsClient)
	wsClientsMu sync.Mutex
)

// BroadcastToStaff ส่ง notification เฉพาะ staff_id ที่ระบุ
func BroadcastToStaff(staffID string, title string, message string) {
	payload, _ := json.Marshal(map[string]string{
		"title":   title,
		"message": message,
	})
	wsClientsMu.Lock()
	defer wsClientsMu.Unlock()
	for conn, cl := range wsClients {
		if cl.staffID == staffID {
			if err := conn.WriteMessage(1, payload); err != nil {
				log.Println("WebSocket send error:", err)
				conn.Close()
				delete(wsClients, conn)
			}
		}
	}
}

func WsHandler(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func WsConnect(c *websocket.Conn) {
	defer c.Close()

	cl := &wsClient{conn: c}
	wsClientsMu.Lock()
	wsClients[c] = cl
	wsClientsMu.Unlock()

	log.Printf("WebSocket เชื่อมต่อ รวม %d คน", len(wsClients))

	for {
		mt, message, err := c.ReadMessage()
		if err != nil || mt == websocket.CloseMessage {
			wsClientsMu.Lock()
			delete(wsClients, c)
			wsClientsMu.Unlock()
			log.Printf("WebSocket หลุด เหลือ %d คน", len(wsClients))
			break
		}

		// รับ {"type":"register","staff_id":"xxx"} จาก client เพื่อผูก staff_id
		var data map[string]string
		if json.Unmarshal(message, &data) == nil {
			if data["type"] == "register" && data["staff_id"] != "" {
				wsClientsMu.Lock()
				cl.staffID = data["staff_id"]
				wsClientsMu.Unlock()
				log.Printf("WebSocket ลงทะเบียน staff_id: %s", cl.staffID)
			}
		}
	}
}
