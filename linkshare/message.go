package linkshare

// Message types are sent to and recevied from clients
type Message struct {
	// Target is the link being shared
	Target string
	// From is the user that sent the link
	From string
	// When is when the link was sent (ISO date)
	When string
}

// Reply sends a message to a client
type Reply func(Message)

// MessageHandler deals with messages from remote clients
type MessageHandler interface {
	OnMessage(*Hub, Message, Reply)
}
