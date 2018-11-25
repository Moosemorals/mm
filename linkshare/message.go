package linkshare

import "encoding/json"

// Message types are sent to and recevied from clients
type Message struct {
	// From gives the source of the message
	From string
	// MsgType indicates the type of message
	MsgType string
	// Payload is the unparsed JSON that depends on the type
	Payload json.RawMessage
}

// Reply sends a message to a client
type Reply func(Message)

// MessageHandler deals with messages from remote clients
type MessageHandler interface {
	OnMessage(*Hub, Message, Reply)
}
