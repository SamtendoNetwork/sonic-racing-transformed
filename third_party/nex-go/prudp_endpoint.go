package nex

import (
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/PretendoNetwork/nex-go/v2/types"
)








type PRUDPEndPoint struct {
	Server                            *PRUDPServer
	StreamID                          uint8
	DefaultStreamSettings             *StreamSettings
	Connections                       *MutexMap[string, *PRUDPConnection]
	packetHandlers                    map[uint16]func(pep *PRUDPEndPoint, packet PRUDPPacketInterface)
	packetEventHandlers               map[string][]func(packet PacketInterface)
	connectionEndedEventHandlers      []func(connection *PRUDPConnection)
	errorEventHandlers                []func(err *Error)
	ConnectionIDCounter               *Counter[uint32]
	ServerAccount                     *Account
	AccountDetailsByPID               func(pid types.PID) (*Account, *Error)
	AccountDetailsByUsername          func(username string) (*Account, *Error)
	IsSecureEndPoint                  bool
	CalcRetransmissionTimeoutCallback CalcRetransmissionTimeoutCallback
}



type CalcRetransmissionTimeoutCallback func(rtt float64, sendCount uint32) time.Duration


func (pep *PRUDPEndPoint) RegisterServiceProtocol(protocol ServiceProtocol) {
	protocol.SetEndpoint(pep)
	pep.OnData(protocol.HandlePacket)
}


func (pep *PRUDPEndPoint) RegisterCustomPacketHandler(packetType uint16, handler func(pep *PRUDPEndPoint, packet PRUDPPacketInterface)) {
	pep.packetHandlers[packetType] = handler
}


func (pep *PRUDPEndPoint) OnData(handler func(packet PacketInterface)) {
	pep.on("data", handler)
}


func (pep *PRUDPEndPoint) OnError(handler func(err *Error)) {

	pep.errorEventHandlers = append(pep.errorEventHandlers, handler)
}




func (pep *PRUDPEndPoint) OnDisconnect(handler func(packet PacketInterface)) {
	pep.on("disconnect", handler)
}




func (pep *PRUDPEndPoint) OnConnectionEnded(handler func(connection *PRUDPConnection)) {

	pep.connectionEndedEventHandlers = append(pep.connectionEndedEventHandlers, handler)
}

func (pep *PRUDPEndPoint) on(name string, handler func(packet PacketInterface)) {
	if _, ok := pep.packetEventHandlers[name]; !ok {
		pep.packetEventHandlers[name] = make([]func(packet PacketInterface), 0)
	}

	pep.packetEventHandlers[name] = append(pep.packetEventHandlers[name], handler)
}


func (pep *PRUDPEndPoint) Emit(name string, packet PRUDPPacketInterface) {
	if handlers, ok := pep.packetEventHandlers[name]; ok {
		for _, handler := range handlers {
			handler(packet)
		}
	}
}

func (pep *PRUDPEndPoint) emitConnectionEnded(connection *PRUDPConnection) {
	for _, handler := range pep.connectionEndedEventHandlers {
		handler(connection)
	}
}


func (pep *PRUDPEndPoint) EmitError(err *Error) {
	for _, handler := range pep.errorEventHandlers {
		handler(err)
	}
}



func (pep *PRUDPEndPoint) CleanupConnection(connection *PRUDPConnection) {
	discriminator := fmt.Sprintf("%s-%d-%d", connection.Socket.Address.String(), connection.StreamType, connection.StreamID)

	found := false
	pep.Connections.RunAndDelete(discriminator, func(key string, conn *PRUDPConnection) {
		found = true
	})


	if !found {
		logger.Warningf("Tried to delete connection %v (ID %v) but it doesn't exist!", discriminator, connection.ID)
	}



	connection.cleanup()
}

func (pep *PRUDPEndPoint) processPacket(packet PRUDPPacketInterface, socket *SocketConnection) {
	streamType := packet.SourceVirtualPortStreamType()
	streamID := packet.SourceVirtualPortStreamID()
	discriminator := fmt.Sprintf("%s-%d-%d", socket.Address.String(), streamType, streamID)
	connection := pep.Connections.GetOrSetDefault(discriminator, func() *PRUDPConnection {
		connection := NewPRUDPConnection(socket)
		connection.endpoint = pep
		connection.ID = pep.ConnectionIDCounter.Next()
		connection.DefaultPRUDPVersion = packet.Version()
		connection.StreamType = streamType
		connection.StreamID = streamID
		connection.StreamSettings = pep.DefaultStreamSettings.Copy()
		return connection
	})

	connection.Lock()
	defer connection.Unlock()

	packet.SetSender(connection)

	if packet.HasFlag(constants.PacketFlagAck) || packet.HasFlag(constants.PacketFlagMultiAck) {
		pep.handleAcknowledgment(packet)
		return
	}

	if packetHandler, ok := pep.packetHandlers[packet.Type()]; ok {
		packetHandler(pep, packet)
	} else {
		logger.Warningf("Unhandled packet type %d", packet.Type())
	}
}

func (pep *PRUDPEndPoint) handleAcknowledgment(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)

	if connection.ConnectionState < StateConnected {
		return
	}

	connection.ResetHeartbeat()

	if packet.HasFlag(constants.PacketFlagMultiAck) {
		pep.handleMultiAcknowledgment(packet)
		return
	}

	if packet.Type() == constants.PingPacket {
		if packet.SequenceID() == connection.outgoingPingSequenceIDCounter.Value {
			connection.rtt.Adjust(time.Since(connection.lastSentPingTime))
		}
	} else {
		logger.Infof("[PRUDP-ACK] remote=%s type=%d seq=%d substream=%d flags=%#x",
			connection.Address().String(), packet.Type(), packet.SequenceID(),
			packet.SubstreamID(), packet.Flags())

		slidingWindow := connection.SlidingWindow(packet.SubstreamID())
		slidingWindow.TimeoutManager.AcknowledgePacket(packet.SequenceID())
	}
}

func (pep *PRUDPEndPoint) handleMultiAcknowledgment(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)
	stream := NewByteStreamIn(packet.Payload(), pep.Server.LibraryVersions, pep.ByteStreamSettings())
	sequenceIDs := make([]uint16, 0)
	var baseSequenceID uint16
	var slidingWindow *SlidingWindow

	if packet.SubstreamID() == 1 || packet.Version() == 2 {


		substreamID, _ := stream.ReadUInt8()
		additionalIDsCount, _ := stream.ReadUInt8()
		baseSequenceID, _ = stream.ReadUInt16LE()
		slidingWindow = connection.SlidingWindow(substreamID)

		for i := 0; i < int(additionalIDsCount); i++ {
			additionalID, _ := stream.ReadUInt16LE()
			sequenceIDs = append(sequenceIDs, additionalID)
		}
	} else {



		slidingWindow = connection.SlidingWindow(0)
		baseSequenceID = packet.SequenceID()

		for stream.Remaining() > 0 {
			additionalID, _ := stream.ReadUInt16LE()
			sequenceIDs = append(sequenceIDs, additionalID)
		}
	}



	slidingWindow.TimeoutManager.packets.Each(func(sequenceID uint16, pending PRUDPPacketInterface) bool {
		if sequenceID <= baseSequenceID && !slices.Contains(sequenceIDs, sequenceID) {
			sequenceIDs = append(sequenceIDs, sequenceID)
		}

		return false
	})


	for _, sequenceID := range sequenceIDs {
		slidingWindow.TimeoutManager.AcknowledgePacket(sequenceID)
	}
}

func (pep *PRUDPEndPoint) handleSyn(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)
	connection.ResetHeartbeat()

	var ack PRUDPPacketInterface

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(pep.Server, connection, nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(pep.Server, connection, nil)
	} else {
		ack, _ = NewPRUDPPacketV0(pep.Server, connection, nil)
	}

	connectionSignature, err := packet.CalculateConnectionSignature(connection.Socket.Address)
	if err != nil {
		logger.Error(err.Error())
	}

	connection.Reset()
	connection.Signature = connectionSignature

	ack.SetType(constants.SynPacket)
	ack.AddFlag(constants.PacketFlagAck)
	ack.AddFlag(constants.PacketFlagHasSize)
	ack.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	ack.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	ack.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	ack.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	ack.SetConnectionSignature(connectionSignature)

	if ack, ok := ack.(*PRUDPPacketV1); ok {

		ack.MaximumSubstreamID = packet.(*PRUDPPacketV1).MaximumSubstreamID
		ack.MinorVersion = packet.(*PRUDPPacketV1).MinorVersion
		ack.SupportedFunctions = pep.Server.SupportedFunctions & packet.(*PRUDPPacketV1).SupportedFunctions
	}

	ack.SetSignature(ack.CalculateSignature([]byte{}, []byte{}))

	connection.ConnectionState = StateConnecting

	pep.Emit("syn", ack)

	pep.Server.SendRaw(connection.Socket, ack.Bytes())
}

func (pep *PRUDPEndPoint) handleConnect(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)

	if connection.ConnectionState < StateConnecting {
		return
	}

	connection.ResetHeartbeat()

	var ack PRUDPPacketInterface

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(pep.Server, connection, nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(pep.Server, connection, nil)
	} else {
		ack, _ = NewPRUDPPacketV0(pep.Server, connection, nil)
	}

	connection.ServerConnectionSignature = packet.GetConnectionSignature()
	connection.SessionID = packet.SessionID()

	connectionSignature, err := packet.CalculateConnectionSignature(connection.Socket.Address)
	if err != nil {
		logger.Error(err.Error())
	}

	connection.ServerSessionID = packet.SessionID()

	ack.SetType(constants.ConnectPacket)
	ack.AddFlag(constants.PacketFlagAck)
	ack.AddFlag(constants.PacketFlagHasSize)
	ack.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	ack.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	ack.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	ack.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	ack.SetConnectionSignature(make([]byte, len(connectionSignature)))
	ack.SetSessionID(connection.ServerSessionID)
	ack.SetSequenceID(1)

	if ack, ok := ack.(*PRUDPPacketV1); ok {




		ack.MaximumSubstreamID = packet.(*PRUDPPacketV1).MaximumSubstreamID
		ack.MinorVersion = packet.(*PRUDPPacketV1).MinorVersion
		ack.SupportedFunctions = packet.(*PRUDPPacketV1).SupportedFunctions

		connection.InitializeSlidingWindows(ack.MaximumSubstreamID)
		connection.InitializePacketDispatchQueues(ack.MaximumSubstreamID)
		connection.OutgoingUnreliableSequenceIDCounter = NewCounter[uint16](packet.(*PRUDPPacketV1).InitialUnreliableSequenceID)
	} else {
		connection.InitializeSlidingWindows(0)
		connection.InitializePacketDispatchQueues(0)
	}

	payload := make([]byte, 0)

	if pep.IsSecureEndPoint {
		var decryptedPayload []byte
		if pep.Server.PRUDPV0Settings.EncryptedConnect {
			decryptedPayload, err = connection.StreamSettings.EncryptionAlgorithm.Decrypt(packet.Payload())
			if err != nil {
				logger.Error(err.Error())
				return
			}

		} else {
			decryptedPayload = packet.Payload()
		}

		decompressedPayload, err := connection.StreamSettings.CompressionAlgorithm.Decompress(decryptedPayload)
		if err != nil {
			logger.Error(err.Error())
			return
		}

		sessionKey, pid, checkValue, err := pep.ReadKerberosTicket(decompressedPayload)
		if err != nil {
			logger.Error(err.Error())
			return
		}

		connection.SetPID(pid)
		connection.SetSessionKey(sessionKey)

		responseCheckValue := checkValue + 1
		responseCheckValueBytes := make([]byte, 4)

		binary.LittleEndian.PutUint32(responseCheckValueBytes, responseCheckValue)

		checkValueResponse := types.NewBuffer(responseCheckValueBytes)
		stream := NewByteStreamOut(pep.Server.LibraryVersions, pep.ByteStreamSettings())

		checkValueResponse.WriteTo(stream)

		payload = stream.Bytes()
	}

	if len(payload) != 0 {
		compressedPayload, err := connection.StreamSettings.CompressionAlgorithm.Compress(payload)
		if err != nil {
			logger.Error(err.Error())
			return
		}

		var encryptedPayload []byte
		if pep.Server.PRUDPV0Settings.EncryptedConnect {
			encryptedPayload, err = connection.StreamSettings.EncryptionAlgorithm.Encrypt(compressedPayload)
			if err != nil {
				logger.Error(err.Error())
				return
			}
		} else {
			encryptedPayload = compressedPayload
		}

		ack.SetPayload(encryptedPayload)
	}

	ack.SetSignature(ack.CalculateSignature([]byte{}, packet.GetConnectionSignature()))

	connection.ConnectionState = StateConnected
	connection.StartHeartbeat()

	pep.Emit("connect", ack)

	pep.Server.SendRaw(connection.Socket, ack.Bytes())
}

func (pep *PRUDPEndPoint) handleData(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)

	if connection.ConnectionState < StateConnected {
		return
	}

	connection.ResetHeartbeat()

	if packet.HasFlag(constants.PacketFlagReliable) {
		pep.HandleReliable(packet)
	} else {
		pep.HandleUnreliable(packet)
	}
}

func (pep *PRUDPEndPoint) handleDisconnect(packet PRUDPPacketInterface) {


	if packet.HasFlag(constants.PacketFlagNeedsAck) {
		pep.AcknowledgePacket(packet)
	}

	streamType := packet.SourceVirtualPortStreamType()
	streamID := packet.SourceVirtualPortStreamID()
	discriminator := fmt.Sprintf("%s-%d-%d", packet.Sender().Address().String(), streamType, streamID)
	if connection, ok := pep.Connections.Get(discriminator); ok {
		pep.CleanupConnection(connection)
	}

	pep.Emit("disconnect", packet)
}

func (pep *PRUDPEndPoint) handlePing(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)

	if connection.ConnectionState < StateConnected {
		return
	}

	connection.ResetHeartbeat()

	if packet.HasFlag(constants.PacketFlagNeedsAck) {
		pep.AcknowledgePacket(packet)
	}

	if packet.HasFlag(constants.PacketFlagReliable) {
		substreamID := packet.SubstreamID()
		packetDispatchQueue := connection.PacketDispatchQueue(substreamID)
		packetDispatchQueue.Queue(packet)
	}
}


func (pep *PRUDPEndPoint) ReadKerberosTicket(payload []byte) ([]byte, types.PID, uint32, error) {
	stream := NewByteStreamIn(payload, pep.Server.LibraryVersions, pep.ByteStreamSettings())

	ticketData := types.NewBuffer(nil)
	if err := ticketData.ExtractFrom(stream); err != nil {
		return nil, 0, 0, err
	}

	requestData := types.NewBuffer(nil)
	if err := requestData.ExtractFrom(stream); err != nil {
		return nil, 0, 0, err
	}


	serverAccount, _ := pep.AccountDetailsByUsername(pep.ServerAccount.Username)
	if serverAccount == nil {
		return nil, 0, 0, errors.New("Failed to find endpoint server account")
	}

	if serverAccount.Password != pep.ServerAccount.Password {
		return nil, 0, 0, errors.New("Password for endpoint server account does not match the records from AccountDetailsByUsername")
	}

	serverKey := DeriveKerberosKey(serverAccount.PID, []byte(serverAccount.Password))

	ticket := NewKerberosTicketInternalData(pep.Server)
	if err := ticket.Decrypt(NewByteStreamIn(ticketData, pep.Server.LibraryVersions, pep.ByteStreamSettings()), serverKey); err != nil {
		return nil, 0, 0, err
	}

	ticketTime := ticket.Issued.Standard()
	serverTime := time.Now().UTC()

	timeLimit := ticketTime.Add(time.Minute * 2)
	if serverTime.After(timeLimit) {
		return nil, 0, 0, errors.New("Kerberos ticket expired")
	}

	sessionKey := ticket.SessionKey
	kerberos := NewKerberosEncryption(sessionKey)

	decryptedRequestData, err := kerberos.Decrypt(requestData)
	if err != nil {
		return nil, 0, 0, err
	}

	checkDataStream := NewByteStreamIn(decryptedRequestData, pep.Server.LibraryVersions, pep.ByteStreamSettings())

	userPID := types.NewPID(0)
	if err := userPID.ExtractFrom(checkDataStream); err != nil {
		return nil, 0, 0, err
	}

	if userPID != ticket.SourcePID {
		return nil, 0, 0, errors.New("User PID and ticket source PID mismatch")
	}

	_, err = checkDataStream.ReadUInt32LE()
	if err != nil {
		return nil, 0, 0, err
	}

	responseCheck, err := checkDataStream.ReadUInt32LE()
	if err != nil {
		return nil, 0, 0, err
	}

	return sessionKey, userPID, responseCheck, nil
}

func (pep *PRUDPEndPoint) AcknowledgePacket(packet PRUDPPacketInterface) {
	var ack PRUDPPacketInterface

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(pep.Server, packet.Sender().(*PRUDPConnection), nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(pep.Server, packet.Sender().(*PRUDPConnection), nil)
	} else {
		ack, _ = NewPRUDPPacketV0(pep.Server, packet.Sender().(*PRUDPConnection), nil)
	}

	ack.SetType(packet.Type())
	ack.AddFlag(constants.PacketFlagAck)
	ack.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	ack.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	ack.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	ack.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	ack.SetSequenceID(packet.SequenceID())
	ack.setFragmentID(packet.getFragmentID())
	ack.SetSubstreamID(packet.SubstreamID())

	pep.Server.sendPacket(ack)


	if packet.Type() == constants.DisconnectPacket {
		pep.Server.sendPacket(ack)
		pep.Server.sendPacket(ack)
	}
}


func (pep *PRUDPEndPoint) HandleReliable(packet PRUDPPacketInterface) {
	if packet.HasFlag(constants.PacketFlagNeedsAck) {
		pep.AcknowledgePacket(packet)
	}

	connection := packet.Sender().(*PRUDPConnection)

	substreamID := packet.SubstreamID()

	packetDispatchQueue := connection.PacketDispatchQueue(substreamID)
	packetDispatchQueue.Queue(packet)

	for nextPacket, ok := packetDispatchQueue.GetNextToDispatch(); ok; nextPacket, ok = packetDispatchQueue.GetNextToDispatch() {
		if nextPacket.Type() == constants.DataPacket {
			var decryptedPayload []byte

			if nextPacket.Version() != 2 {
				decryptedPayload = nextPacket.decryptPayload()
			} else {

				decryptedPayload = nextPacket.Payload()
			}

			decompressedPayload, err := connection.StreamSettings.CompressionAlgorithm.Decompress(decryptedPayload)
			if err != nil {
				logger.Error(err.Error())
			}

			incomingFragmentBuffer := connection.GetIncomingFragmentBuffer(substreamID)
			incomingFragmentBuffer = append(incomingFragmentBuffer, decompressedPayload...)
			connection.SetIncomingFragmentBuffer(substreamID, incomingFragmentBuffer)

			if nextPacket.getFragmentID() == 0 {
				message := NewRMCMessage(pep)
				err := message.FromBytes(incomingFragmentBuffer)
				if err != nil {

					logger.Error(err.Error())
				}

				nextPacket.SetRMCMessage(message)
				connection.ClearOutgoingBuffer(substreamID)

				pep.Emit("data", nextPacket)
			}
		}

		packetDispatchQueue.Dispatched(nextPacket)
	}
}


func (pep *PRUDPEndPoint) HandleUnreliable(packet PRUDPPacketInterface) {
	if packet.HasFlag(constants.PacketFlagNeedsAck) {
		pep.AcknowledgePacket(packet)
	}






































	if packet.getFragmentID() != 0 {
		logger.Warningf("Unexpected unreliable fragment ID. Expected 0, got %d", packet.getFragmentID())
		return
	}

	payload := packet.processUnreliableCrypto()

	message := NewRMCMessage(pep)
	err := message.FromBytes(payload)
	if err != nil {

		logger.Error(err.Error())
	}

	packet.SetRMCMessage(message)

	pep.Emit("data", packet)
}

func (pep *PRUDPEndPoint) sendPing(connection *PRUDPConnection) {
	var ping PRUDPPacketInterface

	switch connection.DefaultPRUDPVersion {
	case 0:
		ping, _ = NewPRUDPPacketV0(pep.Server, connection, nil)
	case 1:
		ping, _ = NewPRUDPPacketV1(pep.Server, connection, nil)
	case 2:
		ping, _ = NewPRUDPPacketLite(pep.Server, connection, nil)
	}

	ping.SetType(constants.PingPacket)
	ping.AddFlag(constants.PacketFlagNeedsAck)
	ping.SetSourceVirtualPortStreamType(connection.StreamType)
	ping.SetSourceVirtualPortStreamID(pep.StreamID)
	ping.SetDestinationVirtualPortStreamType(connection.StreamType)
	ping.SetDestinationVirtualPortStreamID(connection.StreamID)
	ping.SetSubstreamID(0)

	pep.Server.sendPacket(ping)
}


func (pep *PRUDPEndPoint) FindConnectionByID(connectedID uint32) *PRUDPConnection {
	var connection *PRUDPConnection

	pep.Connections.Each(func(discriminator string, pc *PRUDPConnection) bool {
		if pc.ID == connectedID {
			connection = pc
			return true
		}

		return false
	})

	return connection
}


func (pep *PRUDPEndPoint) FindConnectionByPID(pid uint64) *PRUDPConnection {
	var connection *PRUDPConnection

	pep.Connections.Each(func(discriminator string, pc *PRUDPConnection) bool {
		if uint64(pc.pid) == pid && pc.ConnectionState == StateConnected {
			connection = pc
			return true
		}

		return false
	})

	return connection
}


func (pep *PRUDPEndPoint) ComputeRetransmitTimeout(packet PRUDPPacketInterface) time.Duration {
	connection := packet.Sender().(*PRUDPConnection)
	rtt := connection.rtt

	if callback := pep.CalcRetransmissionTimeoutCallback; callback != nil {
		rttAverage := rtt.GetRTTSmoothedAvg()
		rttDeviation := rtt.GetRTTSmoothedDev()
		return callback(rttAverage+rttDeviation*4.0, packet.SendCount())
	}

	var retransmitTimeBase int64
	if packet.Type() == constants.SynPacket {
		retransmitTimeBase = int64(pep.DefaultStreamSettings.SynInitialRTT)
	} else {
		retransmitTimeBase = int64(pep.DefaultStreamSettings.InitialRTT)
		if rtt.Initialized() {
			retransmitTimeBase = int64(rtt.Average()/time.Millisecond) / 8
		}
	}

	retransmitTimeBaseMultiplier := packet.SendCount()

	var retransmitMultiplier float64
	if packet.SendCount() < pep.DefaultStreamSettings.ExtraRetransmitTimeoutTrigger {
		retransmitMultiplier = float64(pep.DefaultStreamSettings.RetransmitTimeoutMultiplier)
	} else {
		retransmitMultiplier = float64(pep.DefaultStreamSettings.ExtraRetransmitTimeoutMultiplier)
	}

	return time.Duration(float64(retransmitTimeBase*int64(retransmitTimeBaseMultiplier))*retransmitMultiplier) * time.Millisecond
}


func (pep *PRUDPEndPoint) AccessKey() string {
	return pep.Server.AccessKey
}


func (pep *PRUDPEndPoint) SetAccessKey(accessKey string) {
	pep.Server.AccessKey = accessKey
}


func (pep *PRUDPEndPoint) Send(packet PacketInterface) {
	pep.Server.Send(packet)
}


func (pep *PRUDPEndPoint) LibraryVersions() *LibraryVersions {
	return pep.Server.LibraryVersions
}


func (pep *PRUDPEndPoint) ByteStreamSettings() *ByteStreamSettings {
	return pep.Server.ByteStreamSettings
}


func (pep *PRUDPEndPoint) SetByteStreamSettings(byteStreamSettings *ByteStreamSettings) {
	pep.Server.ByteStreamSettings = byteStreamSettings
}


func (pep *PRUDPEndPoint) UseVerboseRMC() bool {
	return pep.Server.UseVerboseRMC
}


func (pep *PRUDPEndPoint) EnableVerboseRMC(enable bool) {
	pep.Server.UseVerboseRMC = enable
}


func NewPRUDPEndPoint(streamID uint8) *PRUDPEndPoint {
	pep := &PRUDPEndPoint{
		StreamID:                     streamID,
		DefaultStreamSettings:        NewStreamSettings(),
		Connections:                  NewMutexMap[string, *PRUDPConnection](),
		packetHandlers:               make(map[uint16]func(pep *PRUDPEndPoint, packet PRUDPPacketInterface)),
		packetEventHandlers:          make(map[string][]func(PacketInterface)),
		connectionEndedEventHandlers: make([]func(connection *PRUDPConnection), 0),
		errorEventHandlers:           make([]func(err *Error), 0),
		ConnectionIDCounter:          NewCounter[uint32](0),
		IsSecureEndPoint:             false,
	}

	pep.packetHandlers[constants.SynPacket] = (*PRUDPEndPoint).handleSyn
	pep.packetHandlers[constants.ConnectPacket] = (*PRUDPEndPoint).handleConnect
	pep.packetHandlers[constants.DataPacket] = (*PRUDPEndPoint).handleData
	pep.packetHandlers[constants.DisconnectPacket] = (*PRUDPEndPoint).handleDisconnect
	pep.packetHandlers[constants.PingPacket] = (*PRUDPEndPoint).handlePing

	return pep
}
