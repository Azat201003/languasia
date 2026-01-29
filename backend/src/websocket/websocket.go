package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/Azat201003/languasia/backend/src/database"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	mu     sync.Mutex
	closed bool
	user   database.User
}

type WebSocketHub struct {
	clients    map[uint64]*Client // by user_id
	broadcast  chan Broadcast
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

type Broadcast struct {
	ChatId  uint64 `json:"chat_id"`
	UserId 	uint64 `json:"user_id"`
	Content string `json:"content"`
}

type Message struct {
	Type      string `json:"type"`
}

type ChatMessage struct {
	ChatId 		uint64 `json:"chat_id"`
	Content   string `json:"content"`
}

func NewHub() *WebSocketHub {
	fmt.Println("Creating new hub")
	return &WebSocketHub{
		clients:    make(map[uint64]*Client),
		broadcast:  make(chan Broadcast, 256),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
	}
}

func (h *WebSocketHub) Run() {
	fmt.Println("Hub started running")
	defer fmt.Println("Hub stopped running")

	for {
		select {
		case client := <-h.register:
			fmt.Printf("[HUB] Registering client: %v\n", client)
			h.mutex.Lock()
			h.clients[client.user.UserId] = client
			h.mutex.Unlock()
			fmt.Printf("[HUB] Client registered. Total clients: %d\n", len(h.clients))

		case client := <-h.unregister:
			fmt.Printf("[HUB] Unregistering client: %v (username: %s)\n", client, client.user.Username)
			h.mutex.Lock()
			if h.clients[client.user.UserId] != nil {
				h.clients[client.user.UserId] = nil
				client.mu.Lock()
				if !client.closed {
					client.closed = true
					close(client.send)
				}
				client.mu.Unlock()
				fmt.Printf("[HUB] Client unregistered. Total clients: %d\n", len(h.clients))
			} else {
				fmt.Printf("[HUB] Client not found during unregister: %v\n", client)
			}
			h.mutex.Unlock()

		case broadcast := <-h.broadcast:
			h.mutex.RLock()
			fmt.Printf("[HUB] Starting broadcast to chat with chat_id: %d\n", broadcast.ChatId)
			var userIds []uint64
			if broadcast.ChatId == 0 {
				for userId, _ := range h.clients {
					userIds = append(userIds, userId)
				}
			} else {
				var err error
				userIds, err = database.DBC.GetChatMembers(broadcast.ChatId)

				if err != nil {
					fmt.Printf("[HUB] Broadcasting failed: %v\n", err.Error())
				}
			}

			bytes, err := json.Marshal(broadcast)
			
			if err != nil {
				continue
			}

			for _, id := range userIds {
				client := h.clients[id]
				client.mu.Lock()
				if client.closed {
					h.unregister <- client
					client.mu.Unlock()
					continue
				}

				select{
				case client.send <- bytes:
					fmt.Printf("[HUB] Message sent to client %s\n", client.user.Username)
				default:
					fmt.Printf("[HUB] Client %s send channel full or closed, marking for removal\n", client.user.Username)
					client.closed = true
					h.unregister <- client
				}
				client.mu.Unlock()
			}
			h.mutex.RUnlock()
			
			fmt.Println("[HUB] Broadcast completed")
		}
	}
}

func (h *WebSocketHub) broadcastSystemMessage(content string) {
	fmt.Printf("[HUB] Broadcasting system message: %s\n", content)
	msg := ChatMessage{
		Content:   content,
	}
	bytes, _ := json.Marshal(msg)
	h.broadcast <- Broadcast{0, 0, string(bytes)}
}

func (c *Client) readPump(hub *WebSocketHub) {
	fmt.Printf("[READ_PUMP] Starting read pump for client\n")
	defer func() {
		fmt.Printf("[READ_PUMP] Exiting read pump for client: %s\n", c.user.Username)
		hub.unregister <- c
		c.mu.Lock()
		if !c.closed {
			c.closed = true
			c.conn.Close()
		}
		c.mu.Unlock()
	}()

	// Настройка обработчиков ping/pong
	c.conn.SetPongHandler(func(appData string) error {
		fmt.Printf("[READ_PUMP] Received pong from %s\n", c.user.Username)
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		fmt.Printf("[READ_PUMP] Waiting for message from %s\n", c.user.Username)
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("[READ_PUMP] Unexpected close error from %s: %v\n", c.user.Username, err)
			} else {
				fmt.Printf("[READ_PUMP] Read error from %s: %v (type: %v)\n", c.user.Username, err, messageType)
			}
			return
		}

		fmt.Printf("[READ_PUMP] Received message type %d from %s: %s\n", messageType, c.user.Username, string(message))

		// Парсим сообщение
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			fmt.Printf("[READ_PUMP] Error parsing message from %s: %v\n", c.user.Username, err)
			continue
		}

		// Обрабатываем разные типы сообщений
		switch msg.Type {
		case "chat":
			fmt.Printf("[READ_PUMP] Chat message from %s\n", c.user.Username)
			// Рассылаем сообщение всем клиентам
			var chatMessage ChatMessage
			if err := json.Unmarshal(message, &chatMessage); err != nil {
				fmt.Printf("[READ_PUMP] Error parsing message from %s: %v 2\n", c.user.Username, err)
				continue
			}
			hub.broadcast <- Broadcast{chatMessage.ChatId, c.user.UserId, chatMessage.Content}

		case "ping":
			fmt.Printf("[READ_PUMP] Ping received from %s, sending pong\n", c.user.Username)
			// Отправляем pong в ответ
			pongMsg := Message{
				Type:      "pong",
			}
			pongBytes, _ := json.Marshal(pongMsg)

			c.mu.Lock()
			if !c.closed {
				c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.conn.WriteMessage(websocket.TextMessage, pongBytes); err != nil {
					fmt.Printf("[READ_PUMP] Failed to send pong to %s: %v\n", c.user.Username, err)
					c.mu.Unlock()
					return
				}
			}
			c.mu.Unlock()

		default:
			fmt.Printf("[READ_PUMP] Unknown message type from %s: %s\n", c.user.Username, msg.Type)
		}

		// Сбрасываем дедлайн для следующего чтения
		c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	}
}

func (c *Client) writePump(hub *WebSocketHub) {
	fmt.Printf("[WRITE_PUMP] Starting write pump for client\n")

	// Таймер для отправки ping сообщений
	pingTicker := time.NewTicker(5 * time.Second)
	defer func() {
		fmt.Printf("[WRITE_PUMP] Stopping write pump for client: %s\n", c.user.Username)
		pingTicker.Stop()
		hub.unregister <- c
		c.mu.Lock()
		if !c.closed {
			c.closed = true
			c.conn.Close()
		}
		c.mu.Unlock()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.mu.Lock()
			if c.closed {
				c.mu.Unlock()
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if !ok {
				// Канал закрыт
				fmt.Printf("[WRITE_PUMP] Send channel closed for %s, sending close message\n", c.user.Username)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.mu.Unlock()
				return
			}

			// Пишем сообщение в WebSocket
			fmt.Printf("[WRITE_PUMP] Sending message to %s: %s\n", c.user.Username, string(message))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				fmt.Printf("[WRITE_PUMP] Write error to %s: %v\n", c.user.Username, err)
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()

		case <-pingTicker.C:
			c.mu.Lock()
			if c.closed {
				c.mu.Unlock()
				return
			}

			fmt.Printf("[WRITE_PUMP] Sending ping to %s\n", c.user.Username)
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			// Отправляем ping сообщение
			pingMsg := Message{
				Type:      "ping",
			}
			pingBytes, _ := json.Marshal(pingMsg)

			if err := c.conn.WriteControl(websocket.PingMessage, pingBytes, time.Now().Add(time.Second*2)); err != nil {
				fmt.Printf("[WRITE_PUMP] Failed to send ping to %s: %v\n", c.user.Username, err)
				c.mu.Unlock()
				return
			}

			// Устанавливаем дедлайн для получения pong
			c.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			c.mu.Unlock()
		}
	}
}

func (hub *WebSocketHub) ConnectWebSocket(w http.ResponseWriter, r *http.Request, user database.User) error {
	fmt.Println("[WS] New connection attempt")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[WS] Failed to upgrade connection: %v\n", err)
		return err
	}

	fmt.Println("[WS] Connection upgraded to WebSocket")

	client := &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		closed: false,
		user: user, 
	}

	hub.register <- client
	fmt.Println("[WS] Client registered, starting pumps")

	go client.readPump(hub)
	go client.writePump(hub)

	return nil
}
