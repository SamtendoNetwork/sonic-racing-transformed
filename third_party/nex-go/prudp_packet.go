package nex

import (
	"crypto/rc4"
	"time"

	"github.com/PretendoNetwork/nex-go/v2/constants"
)


type PRUDPPacket struct {
	server                 *PRUDPServer
	sender                 *PRUDPConnection
	readStream             *ByteStreamIn
	version                uint8
	sourceVirtualPort      VirtualPort
	destinationVirtualPort VirtualPort
	packetType             uint16
	flags                  uint16
	sessionID              uint8
	substreamID            uint8
	signature              []byte
	sequenceID             uint16
	connectionSignature    []byte
	fragmentID             uint8
	payload                []byte
	message                *RMCMessage
	sendCount              uint32
	sentAt                 time.Time
	timeout                *Timeout
}


func (p *PRUDPPacket) SetSender(sender ConnectionInterface) {
	p.sender = sender.(*PRUDPConnection)
}


func (p *PRUDPPacket) Sender() ConnectionInterface {
	return p.sender
}


func (p *PRUDPPacket) Flags() uint16 {
	return p.flags
}


func (p *PRUDPPacket) HasFlag(flag uint16) bool {
	return p.flags&flag != 0
}


func (p *PRUDPPacket) AddFlag(flag uint16) {
	p.flags |= flag
}


func (p *PRUDPPacket) SetType(packetType uint16) {
	p.packetType = packetType
}


func (p *PRUDPPacket) Type() uint16 {
	return p.packetType
}


func (p *PRUDPPacket) SetSourceVirtualPortStreamType(streamType constants.StreamType) {
	p.sourceVirtualPort.SetStreamType(streamType)
}


func (p *PRUDPPacket) SourceVirtualPortStreamType() constants.StreamType {
	return p.sourceVirtualPort.StreamType()
}


func (p *PRUDPPacket) SetSourceVirtualPortStreamID(port uint8) {
	p.sourceVirtualPort.SetStreamID(port)
}


func (p *PRUDPPacket) SourceVirtualPortStreamID() uint8 {
	return p.sourceVirtualPort.StreamID()
}


func (p *PRUDPPacket) SetDestinationVirtualPortStreamType(streamType constants.StreamType) {
	p.destinationVirtualPort.SetStreamType(streamType)
}


func (p *PRUDPPacket) DestinationVirtualPortStreamType() constants.StreamType {
	return p.destinationVirtualPort.StreamType()
}


func (p *PRUDPPacket) SetDestinationVirtualPortStreamID(port uint8) {
	p.destinationVirtualPort.SetStreamID(port)
}


func (p *PRUDPPacket) DestinationVirtualPortStreamID() uint8 {
	return p.destinationVirtualPort.StreamID()
}


func (p *PRUDPPacket) SessionID() uint8 {
	return p.sessionID
}


func (p *PRUDPPacket) SetSessionID(sessionID uint8) {
	p.sessionID = sessionID
}


func (p *PRUDPPacket) SubstreamID() uint8 {
	return p.substreamID
}


func (p *PRUDPPacket) SetSubstreamID(substreamID uint8) {
	p.substreamID = substreamID
}

func (p *PRUDPPacket) SetSignature(signature []byte) {
	p.signature = signature
}


func (p *PRUDPPacket) SequenceID() uint16 {
	return p.sequenceID
}


func (p *PRUDPPacket) SetSequenceID(sequenceID uint16) {
	p.sequenceID = sequenceID
}


func (p *PRUDPPacket) Payload() []byte {
	return p.payload
}


func (p *PRUDPPacket) SetPayload(payload []byte) {
	p.payload = payload
}

func (p *PRUDPPacket) decryptPayload() []byte {
	payload := p.payload


	if p.packetType == constants.DataPacket {
		slidingWindow := p.sender.SlidingWindow(p.SubstreamID())

		payload, _ = slidingWindow.streamSettings.EncryptionAlgorithm.Decrypt(payload)
	}

	return payload
}

func (p *PRUDPPacket) GetConnectionSignature() []byte {
	return p.connectionSignature
}

func (p *PRUDPPacket) SetConnectionSignature(connectionSignature []byte) {
	p.connectionSignature = connectionSignature
}

func (p *PRUDPPacket) getFragmentID() uint8 {
	return p.fragmentID
}

func (p *PRUDPPacket) setFragmentID(fragmentID uint8) {
	p.fragmentID = fragmentID
}


func (p *PRUDPPacket) RMCMessage() *RMCMessage {
	return p.message
}


func (p *PRUDPPacket) SetRMCMessage(message *RMCMessage) {
	p.message = message
}


func (p *PRUDPPacket) SendCount() uint32 {
	return p.sendCount
}

func (p *PRUDPPacket) incrementSendCount() {
	p.sendCount++
}


func (p *PRUDPPacket) SentAt() time.Time {
	return p.sentAt
}

func (p *PRUDPPacket) setSentAt(time time.Time) {
	p.sentAt = time
}

func (p *PRUDPPacket) getTimeout() *Timeout {
	return p.timeout
}

func (p *PRUDPPacket) setTimeout(timeout *Timeout) {
	p.timeout = timeout
}

func (p *PRUDPPacket) processUnreliableCrypto() []byte {


	uniqueKey := p.sender.UnreliablePacketBaseKey[:]
	uniqueKey[0] = byte((uint16(uniqueKey[0]) + p.sequenceID) & 0xFF)
	uniqueKey[1] = byte((uint16(uniqueKey[1]) + (p.sequenceID >> 8)) & 0xFF)
	uniqueKey[31] = byte((uniqueKey[31] + p.sessionID) & 0xFF)

	cipher, _ := rc4.NewCipher(uniqueKey)
	ciphered := make([]byte, len(p.payload))

	cipher.XORKeyStream(ciphered, p.payload)

	return ciphered
}
