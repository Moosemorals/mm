package linkshare

// Adapted from https://github.com/gorilla/websocket/blob/master/examples/chat

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// A client talks to the websocket and the hub
type client struct {
	hub  *Hub
	conn *websocket.Conn
	out  chan Message
}

func (c *client) setReadDeadline() {
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *client) setWriteDeadline() {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
}

// Read messages from the client to the hub
func (c *client) read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.setReadDeadline()
	c.conn.SetPongHandler(func(string) error {
		c.setReadDeadline()
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Websocket client error: %v", err)
			}
			break
		}
		c.hub.in <- clientMessage{c, msg}
	}
}

// Write messages from the hub to the socket
func (c *client) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.out:
			c.setWriteDeadline()
			if !ok {
				// the hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteJSON(msg)

			// Write any queued messages as well
			n := len(c.out)
			for i := 0; i < n; i++ {
				err := c.conn.WriteJSON(<-c.out)
				if err != nil {
					return
				}
			}
		case <-ticker.C:
			c.setWriteDeadline()
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}

type clientMessage struct {
	c *client
	m Message
}

// Hub manages clients and broadcasts messages
type Hub struct {
	clients    map[*client]bool
	handlers   map[MessageHandler]bool
	in         chan clientMessage
	register   chan *client
	unregister chan *client
	shutdown   chan int
}

// NewHub initilises a new Hub
func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[*client]bool),
		handlers:   make(map[MessageHandler]bool),
		in:         make(chan clientMessage),
		register:   make(chan *client),
		unregister: make(chan *client),
		shutdown:   make(chan int),
	}
	go h.run()
	return h
}

// Broadcast a message to all listening clients
func (h *Hub) Broadcast(m Message) {
	for client := range h.clients {
		select {
		case client.out <- m:
		default:
			h.closeclient(client)
		}
	}
}

// RegisterHandler to handle messages
func (h *Hub) RegisterHandler(mh MessageHandler) {
	h.handlers[mh] = true
}

// UnregisterHandler to stop handling messages
func (h *Hub) UnregisterHandler(mh MessageHandler) {
	delete(h.handlers, mh)
}

// Shutdown client websockets cleanly
func (h *Hub) Shutdown() {
	h.shutdown <- 1
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrading websocket failed: %v", err)
		return
	}

	client := &client{
		hub:  h,
		conn: conn,
		out:  make(chan Message, 5),
	}

	h.register <- client

	// Start client goroutines
	go client.write()
	go client.read()
}

// Helper function, closes a client connection
func (h *Hub) closeclient(c *client) {
	close(c.out)
	delete(h.clients, c)
}

// Main loop. Wait for messages from clients,
// hand them off to handlers, and deal with
// replies
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.closeclient(client)
			}
		case cm := <-h.in:
			for handler := range h.handlers {
				handler.OnMessage(h, cm.m, func(m Message) {
					cm.c.out <- m
				})
			}
		case <-h.shutdown:
			log.Printf("Closing %d websockets", len(h.clients))
			for client := range h.clients {
				h.closeclient(client)
			}
			return
		}
	}
}
