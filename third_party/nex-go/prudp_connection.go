package nex

import (
	"crypto/md5"
	"net"
	"sync"
	"time"

	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/PretendoNetwork/nex-go/v2/types"
)




type PRUDPConnection struct {
	Socket                              *SocketConnection
	endpoint                            *PRUDPEndPoint
	ConnectionState                     ConnectionState
	ID                                  uint32
	SessionID                           uint8
	ServerSessionID                     uint8
	SessionKey                          []byte
	pid                                 types.PID
	DefaultPRUDPVersion                 int
	StreamType                          constants.StreamType
	StreamID                            uint8
	StreamSettings                      *StreamSettings
	Signature                           []byte
	ServerConnectionSignature           []byte
	UnreliablePacketBaseKey             []byte
	rtt                                 *RTT
	slidingWindows                      *MutexMap[uint8, *SlidingWindow]
	packetDispatchQueues                *MutexMap[uint8, *PacketDispatchQueue]
	incomingFragmentBuffers             *MutexMap[uint8, []byte]
	OutgoingUnreliableSequenceIDCounter *Counter[uint16]
	outgoingPingSequenceIDCounter       *Counter[uint16]
	lastSentPingTime                    time.Time
	heartbeatTimer                      *time.Timer
	pingKickTimer                       *time.Timer
	StationURLs                         types.List[types.StationURL]
	mutex                               *sync.Mutex
}


func (pc *PRUDPConnection) Endpoint() EndpointInterface {
	return pc.endpoint
}


func (pc *PRUDPConnection) Address() net.Addr {
	return pc.Socket.Address
}


func (pc *PRUDPConnection) PID() types.PID {
	return pc.pid
}


func (pc *PRUDPConnection) SetPID(pid types.PID) {
	pc.pid = pid
}


func (pc *PRUDPConnection) Reset() {
	pc.ConnectionState = StateNotConnected
	pc.packetDispatchQueues.Clear(func(_ uint8, packetDispatchQueue *PacketDispatchQueue) {
		packetDispatchQueue.Purge()
	})

	pc.slidingWindows.Clear(func(_ uint8, slidingWindow *SlidingWindow) {
		slidingWindow.TimeoutManager.Stop()
	})

	pc.Signature = make([]byte, 0)
	pc.ServerConnectionSignature = make([]byte, 0)
	pc.SessionKey = make([]byte, 0)
	pc.OutgoingUnreliableSequenceIDCounter = NewCounter[uint16](1)
	pc.outgoingPingSequenceIDCounter = NewCounter[uint16](0)
}


func (pc *PRUDPConnection) cleanup() {
	pc.Reset()

	pc.stopHeartbeatTimers()

	pc.endpoint.emitConnectionEnded(pc)
}


func (pc *PRUDPConnection) InitializeSlidingWindows(maxSubstreamID uint8) {

	pc.slidingWindows = NewMutexMap[uint8, *SlidingWindow]()

	for i := 0; i < int(maxSubstreamID+1); i++ {
		pc.CreateSlidingWindow(uint8(i))
	}
}


func (pc *PRUDPConnection) InitializePacketDispatchQueues(maxSubstreamID uint8) {

	pc.packetDispatchQueues = NewMutexMap[uint8, *PacketDispatchQueue]()

	for i := 0; i < int(maxSubstreamID+1); i++ {
		pc.CreatePacketDispatchQueue(uint8(i))
	}
}



func (pc *PRUDPConnection) CreateSlidingWindow(substreamID uint8) *SlidingWindow {
	slidingWindow := NewSlidingWindow()
	slidingWindow.sequenceIDCounter = NewCounter[uint16](0)
	slidingWindow.streamSettings = pc.StreamSettings.Copy()

	pc.slidingWindows.Set(substreamID, slidingWindow)

	return slidingWindow
}


func (pc *PRUDPConnection) SlidingWindow(substreamID uint8) *SlidingWindow {
	slidingWindow, ok := pc.slidingWindows.Get(substreamID)
	if !ok {




		slidingWindow = pc.CreateSlidingWindow(substreamID)
	}

	return slidingWindow
}


func (pc *PRUDPConnection) CreatePacketDispatchQueue(substreamID uint8) *PacketDispatchQueue {
	pdq := NewPacketDispatchQueue()
	pc.packetDispatchQueues.Set(substreamID, pdq)
	return pdq
}



func (pc *PRUDPConnection) PacketDispatchQueue(substreamID uint8) *PacketDispatchQueue {
	packetDispatchQueue, ok := pc.packetDispatchQueues.Get(substreamID)
	if !ok {




		packetDispatchQueue = pc.CreatePacketDispatchQueue(substreamID)
	}

	return packetDispatchQueue
}


func (pc *PRUDPConnection) SetSessionKey(sessionKey []byte) {
	pc.SessionKey = sessionKey

	pc.slidingWindows.Each(func(substreamID uint8, slidingWindow *SlidingWindow) bool {




		if substreamID == 0 {
			slidingWindow.SetCipherKey(sessionKey)
		} else {
			modifier := len(sessionKey)/2 + 1


			sessionKey = append(make([]byte, 0), sessionKey...)


			for i := 0; i < len(sessionKey)/2; i++ {
				sessionKey[i] = (sessionKey[i] + byte(modifier-i)) & 0xFF
			}

			slidingWindow.SetCipherKey(sessionKey)
		}

		return false
	})







	unreliableBaseKeyPart1 := md5.Sum(append(sessionKey, []byte{0x18, 0xD8, 0x23, 0x34, 0x37, 0xE4, 0xE3, 0xFE}...))
	unreliableBaseKeyPart2 := md5.Sum(append(sessionKey, []byte{0x23, 0x3E, 0x60, 0x01, 0x23, 0xCD, 0xAB, 0x80}...))

	pc.UnreliablePacketBaseKey = append(unreliableBaseKeyPart1[:], unreliableBaseKeyPart2[:]...)
}


func (pc *PRUDPConnection) ResetHeartbeat() {
	if pc.pingKickTimer != nil {
		pc.pingKickTimer.Stop()
	}

	if pc.heartbeatTimer != nil {

		pc.heartbeatTimer.Reset(time.Duration(pc.StreamSettings.MaxSilenceTime) * time.Millisecond)
	}
}




func (pc *PRUDPConnection) Lock() {
	pc.mutex.Lock()
}




func (pc *PRUDPConnection) Unlock() {
	pc.mutex.Unlock()
}


func (pc *PRUDPConnection) GetIncomingFragmentBuffer(substreamID uint8) []byte {
	buffer, ok := pc.incomingFragmentBuffers.Get(substreamID)
	if !ok {
		buffer = make([]byte, 0)
		pc.incomingFragmentBuffers.Set(substreamID, buffer)
	}

	return buffer
}


func (pc *PRUDPConnection) SetIncomingFragmentBuffer(substreamID uint8, buffer []byte) {
	pc.incomingFragmentBuffers.Set(substreamID, buffer)
}


func (pc *PRUDPConnection) ClearOutgoingBuffer(substreamID uint8) {
	pc.incomingFragmentBuffers.Set(substreamID, make([]byte, 0))
}


func (pc *PRUDPConnection) StartHeartbeat() {
	endpoint := pc.endpoint


	maxSilenceTime := time.Duration(pc.StreamSettings.MaxSilenceTime) * time.Millisecond







	pc.heartbeatTimer = time.AfterFunc(maxSilenceTime, func() {
		endpoint.sendPing(pc)



		pc.pingKickTimer = time.AfterFunc(maxSilenceTime, func() {
			endpoint.CleanupConnection(pc)
		})
	})
}

func (pc *PRUDPConnection) stopHeartbeatTimers() {
	if pc.pingKickTimer != nil {
		pc.pingKickTimer.Stop()
	}

	if pc.heartbeatTimer != nil {
		pc.heartbeatTimer.Stop()
	}
}


func NewPRUDPConnection(socket *SocketConnection) *PRUDPConnection {
	pc := &PRUDPConnection{
		Socket:                              socket,
		ConnectionState:                     StateNotConnected,
		rtt:                                 NewRTT(),
		pid:                                 types.NewPID(0),
		slidingWindows:                      NewMutexMap[uint8, *SlidingWindow](),
		packetDispatchQueues:                NewMutexMap[uint8, *PacketDispatchQueue](),
		OutgoingUnreliableSequenceIDCounter: NewCounter[uint16](1),
		outgoingPingSequenceIDCounter:       NewCounter[uint16](0),
		incomingFragmentBuffers:             NewMutexMap[uint8, []byte](),
		StationURLs:                         types.NewList[types.StationURL](),
		mutex:                               &sync.Mutex{},
		UnreliablePacketBaseKey:             make([]byte, md5.Size*2),
	}

	return pc
}
