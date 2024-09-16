package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gofiber/contrib/websocket"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The session ID.
	sessionID string
	// The websocket connection.
	conn *websocket.Conn

	topicMu sync.RWMutex
	// List of tops the client is subscribed to.
	topics []structures.Topic

	// Buffered channel of outbound messages.
	send chan []byte

	// Ping mutex
	pingMu sync.Mutex
	// Number of max pings without pong.
	pingCycle uint8
	// Number of pings without pong.
	missedPings uint8
	// Number of pings sent.
	pingCount uint64
}

// Ping sends a ping to the client.
func (c *Client) Ping() {
	err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		log.Debug("failed to set write deadline", "error", err)
	}

	c.pingMu.Lock()
	c.pingCount++
	c.pingMu.Unlock()

	// Send the heartbeat message
	c.SendMessage(structures.NewMessage(structures.CodeHeartbeat, structures.HeartbeatPayload{
		Count: c.pingCount,
	}))
}

// Close closes the client connection and unregisters it from the hub.
func (c *Client) Close() {
	c.conn.Close()
	c.hub.unregister <- c
}

func (c *Client) Subscribe(topic structures.Topic) {
	c.topicMu.Lock()
	defer c.topicMu.Unlock()

	c.topics = append(c.topics, topic)
}

func (c *Client) IsSubscribed(topic structures.Topic) bool {
	c.topicMu.RLock()
	defer c.topicMu.RUnlock()

	for _, t := range c.topics {
		if t == topic {
			return true
		}
	}

	return false
}

func (c *Client) Unsubscribe(topic structures.Topic) {
	c.topicMu.Lock()
	defer c.topicMu.Unlock()

	for i, t := range c.topics {
		if t == topic {
			c.topics = append(c.topics[:i], c.topics[i+1:]...)
			return
		}
	}
}

func (c *Client) Topics() []structures.Topic {
	c.topicMu.RLock()
	defer c.topicMu.RUnlock()

	return c.topics
}

func (c *Client) SendMessage(msg structures.Message) {
	msgPayload, err := json.Marshal(msg)
	if err != nil {
		log.Error("failed to marshal message", "error", err)
		return
	}

	select {
	case c.send <- msgPayload:

	default:
		log.Warn("SendMessage: send channel is full, message dropped", "sessionID", c.sessionID)
	}
}
