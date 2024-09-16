package structures

import (
	"time"
)

type Code uint8

const (
	// Default codes (0/10)
	CodeDispatch  Code = 0 // Dispatches an event to client
	CodeHello     Code = 1 // Sent immediately after connecting, contains session ID
	CodeHeartbeat Code = 2 // Sent by client to keep connection alive
	CodeAck       Code = 3 // Acknowledges a message

	// Command codes (11/20)
	CodeSubscribe   Code = 11 // Subscribe to a topic
	CodeUnsubscribe Code = 12 // Unsubscribe from a topic
)

type Topic string

func (t Topic) String() string {
	return string(t)
}

const (
	TopicJobRan Topic = "job.ran" // Topic for when a job has ran
)

type Message struct {
	Code      Code        `json:"c"`
	Timestamp int64       `json:"t"`
	Data      interface{} `json:"d"`
}

func NewMessage(code Code, data interface{}) Message {
	return Message{
		Code:      code,
		Timestamp: time.Now().UnixMilli(),
		Data:      data,
	}
}

type HelloPayload struct {
	SessionID string `json:"session_id"`
}

// Sent to the client per heartbeat
type HeartbeatPayload struct {
	Count uint64 `json:"count"`
}

type SubscribePayload struct {
	Topics []Topic `json:"topics"`
}

type UnsubscribePayload struct {
	Topics []Topic `json:"topics"`
}

type DispatchPayload struct {
	Topic Topic       `json:"topic"`
	Data  interface{} `json:"data"`
}
