package nex


type PacketInterface interface {
	Sender() ConnectionInterface
	Payload() []byte
	SetPayload(payload []byte)
	RMCMessage() *RMCMessage
	SetRMCMessage(message *RMCMessage)
}
