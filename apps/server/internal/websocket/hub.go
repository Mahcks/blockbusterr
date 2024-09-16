package websocket

import (
	"context"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/services"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// App context
	gctx context.Context

	// Crate of services
	crate *services.Crate

	// Registered clients.
	clients map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Stop requests to stop the hub.
	stop chan struct{}

	// Wait for hub to finish shutdown
	wg sync.WaitGroup
	// Done to signal hub is shutdown
	done chan struct{}
}

func NewHub(gctx global.Context) *Hub {
	return &Hub{
		gctx:       gctx,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	h.wg.Add(1)

	for {
		select {
		// Register a new client
		case client := <-h.register:
			h.clients[client] = true
			log.Infof("client registered: %v", client.sessionID)

		// Unregister a client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Infof("client disconnected: %v", client.sessionID)
			}

		// Stop the hub
		case <-h.stop:
			for client := range h.clients {
				close(client.send)
			}
			h.wg.Done()
			close(h.done)
			return

		case <-h.gctx.Done():
			// TODO: Message all clients a closing message
			h.wg.Done()
			close(h.done)
			return
		}
	}
}

// Stop stops the hub and closes all connections.
func (h *Hub) Stop() {
	// TODO: Send closing message to clients

	h.stop <- struct{}{}
}

func (h *Hub) Wait() {
	<-h.done
}

// ServeWs handles websocket requests from the peer.
func (h *Hub) ServeWs(conn *websocket.Conn) {
	client := &Client{
		sessionID: uuid.NewString(),
		hub:       h,
		conn:      conn,
		send:      make(chan []byte, 256),
		pingCycle: 3,
	}

	client.hub.register <- client
	log.Debug("New connection")

	client.SendMessage(structures.NewMessage(structures.CodeHello, structures.HelloPayload{
		SessionID: client.sessionID,
	}))

	go client.writePump()
	client.readPump()
}

func (h *Hub) SendMessageToTopic(topic structures.Topic, msg structures.Message) {
	if h.clients == nil || len(h.clients) == 0 {
		log.Debug("SendMessageToTopic: no clients connected", "topic", topic)
		return
	}

	for client := range h.clients {
		// Check if the client is subscribed to the topic.
		if client.IsSubscribed(topic) {
			client.SendMessage(msg)
		} else {
			log.Debug("SendMessageToTopic: client is not subscribed to topic", "sessionID", client.sessionID, "topic", topic)
		}
	}
}
