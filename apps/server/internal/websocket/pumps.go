package websocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gofiber/contrib/websocket"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)

	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Debugf("failed to set read deadline: %v", err)
	}

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.conn.SetReadDeadline(time.Now().Add(pongWait))

		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))

		var msg structures.Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Debug("failed to unmarshal message", "error", err)
			return
		}

		log.Debug("[WS] Received message", "message", msg)

		switch msg.Code {
		case structures.CodeHeartbeat:
			c.pingMu.Lock()
			c.missedPings = 0
			c.pingMu.Unlock()

		case structures.CodeSubscribe:
			rawPayload, err := json.Marshal(msg.Data)
			if err != nil {
				log.Debug("failed to unmarshal subscribe payload", "error", err)
				return
			}

			var payload structures.SubscribePayload
			err = json.Unmarshal(rawPayload, &payload)
			if err != nil {
				log.Debug("failed to unmarshal subscribe payload", "error", err)
				return
			}

			fmt.Println("Subscribing to", payload.Topics)

			for _, topic := range payload.Topics {
				c.Subscribe(topic)
			}

		case structures.CodeUnsubscribe:
			rawPayload, err := json.Marshal(msg.Data)
			if err != nil {
				log.Debug("failed to unmarshal unsubscribe payload", "error", err)
				return
			}

			var payload structures.SubscribePayload
			err = json.Unmarshal(rawPayload, &payload)
			if err != nil {
				log.Debug("failed to unmarshal unsubscribe payload", "error", err)
				return
			}

			fmt.Println("Unsubscribing to", payload.Topics)

			for _, topic := range payload.Topics {
				c.Unsubscribe(topic)
			}
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Debug("failed to set write deadline", "error", err)
			}

			if !ok {
				// The hub closed the channel.
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Debug("error writing close message", "error", err)
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			_, err = w.Write(message)
			if err != nil {
				log.Debug("error writing message", "error", err)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Ping()
		}
	}
}
