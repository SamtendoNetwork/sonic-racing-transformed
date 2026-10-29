package nex

import (
	"bytes"
	"crypto/rand"
	"expvar"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/lxzan/gws"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)




func EnableBasicUDPHealthCheck(port int) {
	udpAddress, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	socket, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		panic(err)
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			buffer := make([]byte, 1024)

			for {
				n, clientAddr, err := socket.ReadFromUDP(buffer)
				if err == nil {
					socket.WriteToUDP(buffer[:n], clientAddr)
				}
			}
		}()
	}
}


type PRUDPServer struct {
	udpSocket                     *net.UDPConn
	websocketServer               *WebSocketServer
	Endpoints                     *MutexMap[uint8, *PRUDPEndPoint]
	SupportedFunctions            uint32
	AccessKey                     string
	KerberosTicketVersion         int
	SessionKeyLength              int
	FragmentSize                  int
	PRUDPv1ConnectionSignatureKey []byte
	LibraryVersions               *LibraryVersions
	ByteStreamSettings            *ByteStreamSettings
	PRUDPV0Settings               *PRUDPV0Settings
	PRUDPV1Settings               *PRUDPV1Settings
	UseVerboseRMC                 bool
}




func (ps *PRUDPServer) EnableMetrics(addr string) {
	endpointConnections := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "prudp_endpoint_connections",
			Help: "Number of active connections per PRUDP endpoint",
		},
		[]string{"endpoint_id"},
	)

	totalConnections := promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "prudp_total_connections",
			Help: "Total number of active PRUDP connections across all endpoints",
		},
	)

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			total := 0
			ps.Endpoints.Each(func(key uint8, endpoint *PRUDPEndPoint) bool {
				count := endpoint.Connections.Size()
				total += count
				endpointConnections.WithLabelValues(fmt.Sprintf("prudp_endpoint_%d", endpoint.StreamID)).Set(float64(count))
				return false
			})
			totalConnections.Set(float64(total))
		}
	}()

	expvar.Publish("endpoint_connections", expvar.Func(func() any {
		result := make(map[string]any)

		endpointCounts := make(map[string]int)

		ps.Endpoints.Each(func(_ uint8, endpoint *PRUDPEndPoint) bool {
			endpointCounts[fmt.Sprintf("prudp_endpoint_%d", endpoint.StreamID)] = endpoint.Connections.Size()
			return false
		})

		result["prudp_endpoint_connections"] = endpointCounts

		return result
	}))

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Errorf("pprof server failed: %v", err)
		}
	}()
}


func (ps *PRUDPServer) BindPRUDPEndPoint(endpoint *PRUDPEndPoint) {
	if ps.Endpoints.Has(endpoint.StreamID) {
		logger.Warningf("Tried to bind already existing PRUDPEndPoint %d", endpoint.StreamID)
		return
	}

	endpoint.Server = ps
	ps.Endpoints.Set(endpoint.StreamID, endpoint)
}


func (ps *PRUDPServer) Listen(port int) {
	ps.ListenUDP(port)
}


func (ps *PRUDPServer) ListenUDP(port int) {
	ps.initPRUDPv1ConnectionSignatureKey()

	udpAddress, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	socket, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		panic(err)
	}

	ps.udpSocket = socket

	quit := make(chan struct{})

	for i := 0; i < runtime.NumCPU(); i++ {
		go ps.listenDatagram(quit)
	}

	<-quit
}

func (ps *PRUDPServer) listenDatagram(quit chan struct{}) {
	var err error
	buffer := make([]byte, 64000)

	for err == nil {
		var read int
		var addr *net.UDPAddr

		read, addr, err = ps.udpSocket.ReadFromUDP(buffer)
		if err == nil {
			packetData := make([]byte, read)
			copy(packetData, buffer[:read])

			err = ps.handleSocketMessage(packetData, addr, nil)
		}
	}

	quit <- struct{}{}

	panic(err)
}


func (ps *PRUDPServer) ListenWebSocket(port int) {
	ps.initPRUDPv1ConnectionSignatureKey()

	ps.websocketServer = &WebSocketServer{
		prudpServer: ps,
	}

	ps.websocketServer.listen(port)
}


func (ps *PRUDPServer) ListenWebSocketSecure(port int, certFile, keyFile string) {
	ps.initPRUDPv1ConnectionSignatureKey()

	ps.websocketServer = &WebSocketServer{
		prudpServer: ps,
	}

	ps.websocketServer.listenSecure(port, certFile, keyFile)
}

func (ps *PRUDPServer) initPRUDPv1ConnectionSignatureKey() {

	if len(ps.PRUDPv1ConnectionSignatureKey) != 16 {
		ps.PRUDPv1ConnectionSignatureKey = make([]byte, 16)
		_, err := rand.Read(ps.PRUDPv1ConnectionSignatureKey)
		if err != nil {
			panic(err)
		}
	}
}

func (ps *PRUDPServer) handleSocketMessage(packetData []byte, address net.Addr, webSocketConnection *gws.Conn) error {

	if len(packetData) < 2 {
		return nil
	}

	readStream := NewByteStreamIn(packetData, ps.LibraryVersions, ps.ByteStreamSettings)

	var packets []PRUDPPacketInterface





	if ps.websocketServer != nil && packetData[0] == 0x80 {
		packets, _ = NewPRUDPPacketsLite(ps, nil, readStream)
	} else if bytes.Equal(packetData[:2], []byte{0xEA, 0xD0}) {
		packets, _ = NewPRUDPPacketsV1(ps, nil, readStream)
	} else {
		packets, _ = NewPRUDPPacketsV0(ps, nil, readStream)
	}

	for _, packet := range packets {
		go ps.processPacket(packet, address, webSocketConnection)
	}

	return nil
}

func (ps *PRUDPServer) processPacket(packet PRUDPPacketInterface, address net.Addr, webSocketConnection *gws.Conn) {
	if !ps.Endpoints.Has(packet.DestinationVirtualPortStreamID()) {
		logger.Warningf("Client %s trying to connect to unbound PRUDPEndPoint %d", address.String(), packet.DestinationVirtualPortStreamID())
		return
	}

	endpoint, ok := ps.Endpoints.Get(packet.DestinationVirtualPortStreamID())
	if !ok {
		logger.Warningf("Client %s trying to connect to unbound PRUDPEndPoint %d", address.String(), packet.DestinationVirtualPortStreamID())
		return
	}

	if packet.DestinationVirtualPortStreamType() != packet.SourceVirtualPortStreamType() {
		logger.Warningf("Client %s trying to use non matching destination and source stream types %d and %d", address.String(), packet.DestinationVirtualPortStreamType(), packet.SourceVirtualPortStreamType())
		return
	}

	if packet.DestinationVirtualPortStreamType() > constants.StreamTypeRelay {
		logger.Warningf("Client %s trying to use invalid to destination stream type %d", address.String(), packet.DestinationVirtualPortStreamType())
		return
	}

	if packet.SourceVirtualPortStreamType() > constants.StreamTypeRelay {
		logger.Warningf("Client %s trying to use invalid to source stream type %d", address.String(), packet.DestinationVirtualPortStreamType())
		return
	}

	sourcePortNumber := packet.SourceVirtualPortStreamID()
	invalidSourcePort := false



	if _, ok := packet.(*PRUDPPacketLite); ok {
		if sourcePortNumber > 31 {
			invalidSourcePort = true
		}
	} else if sourcePortNumber > 15 {
		invalidSourcePort = true
	}

	if invalidSourcePort {
		logger.Warningf("Client %s trying to use invalid to source port number %d. Port number too large", address.String(), sourcePortNumber)
		return
	}

	socket := NewSocketConnection(ps, address, webSocketConnection)
	endpoint.processPacket(packet, socket)
}


func (ps *PRUDPServer) Send(packet PacketInterface) {
	if packet, ok := packet.(PRUDPPacketInterface); ok {
		data := packet.Payload()
		fragments := int(len(data) / ps.FragmentSize)

		var fragmentID uint8 = 1
		for i := 0; i <= fragments; i++ {
			if len(data) < ps.FragmentSize {
				packet.SetPayload(data)
				packet.setFragmentID(0)
			} else {
				packet.SetPayload(data[:ps.FragmentSize])
				packet.setFragmentID(fragmentID)

				data = data[ps.FragmentSize:]
				fragmentID++
			}

			ps.sendPacket(packet)





			if i < fragments {
				time.Sleep(16 * time.Millisecond)
			}
		}
	}
}

func (ps *PRUDPServer) sendPacket(packet PRUDPPacketInterface) {




	packetCopy := packet.Copy()
	connection := packetCopy.Sender().(*PRUDPConnection)

	if !packetCopy.HasFlag(constants.PacketFlagAck) && !packetCopy.HasFlag(constants.PacketFlagMultiAck) {
		if packetCopy.HasFlag(constants.PacketFlagReliable) {
			slidingWindow := connection.SlidingWindow(packetCopy.SubstreamID())
			packetCopy.SetSequenceID(slidingWindow.NextOutgoingSequenceID())
		} else if packetCopy.Type() == constants.DataPacket {
			packetCopy.SetSequenceID(connection.OutgoingUnreliableSequenceIDCounter.Next())
		} else if packetCopy.Type() == constants.PingPacket {
			packetCopy.SetSequenceID(connection.outgoingPingSequenceIDCounter.Next())
			connection.lastSentPingTime = time.Now()
		} else {
			packetCopy.SetSequenceID(0)
		}
	}

	packetCopy.SetSessionID(connection.ServerSessionID)

	if packetCopy.Type() == constants.DataPacket && !packetCopy.HasFlag(constants.PacketFlagAck) && !packetCopy.HasFlag(constants.PacketFlagMultiAck) {
		if packetCopy.HasFlag(constants.PacketFlagReliable) {
			slidingWindow := connection.SlidingWindow(packetCopy.SubstreamID())
			payload := packetCopy.Payload()

			if len(payload) != 0 {
				compressedPayload, err := slidingWindow.streamSettings.CompressionAlgorithm.Compress(payload)
				if err != nil {
					logger.Error(err.Error())
				}

				encryptedPayload, err := slidingWindow.streamSettings.EncryptionAlgorithm.Encrypt(compressedPayload)
				if err != nil {
					logger.Error(err.Error())
				}

				packetCopy.SetPayload(encryptedPayload)
			}
		} else {

			if packetCopy.Version() != 2 {
				packetCopy.SetPayload(packetCopy.processUnreliableCrypto())
			}
		}
	}

	if ps.PRUDPV1Settings.LegacyConnectionSignature {
		packetCopy.SetSignature(packetCopy.CalculateSignature(connection.SessionKey, connection.Signature))
	} else {
		packetCopy.SetSignature(packetCopy.CalculateSignature(connection.SessionKey, connection.ServerConnectionSignature))
	}

	packetCopy.incrementSendCount()
	packetCopy.setSentAt(time.Now())

	if packetCopy.HasFlag(constants.PacketFlagReliable) && packetCopy.HasFlag(constants.PacketFlagNeedsAck) {
		slidingWindow := connection.SlidingWindow(packetCopy.SubstreamID())
		slidingWindow.TimeoutManager.SchedulePacketTimeout(packetCopy)
	}

	ps.SendRaw(packetCopy.Sender().(*PRUDPConnection).Socket, packetCopy.Bytes())
}


func (ps *PRUDPServer) SendRaw(socket *SocketConnection, data []byte) {


	var err error

	if address, ok := socket.Address.(*net.UDPAddr); ok && ps.udpSocket != nil {
		_, err = ps.udpSocket.WriteToUDP(data, address)
	} else if socket.WebSocketConnection != nil {
		err = socket.WebSocketConnection.WriteMessage(gws.OpcodeBinary, data)
	}

	if err != nil {
		logger.Error(err.Error())
	}
}


func (ps *PRUDPServer) SetFragmentSize(fragmentSize int) {











	ps.FragmentSize = fragmentSize
}


func NewPRUDPServer() *PRUDPServer {
	return &PRUDPServer{
		Endpoints:          NewMutexMap[uint8, *PRUDPEndPoint](),
		SessionKeyLength:   32,
		FragmentSize:       1300,
		LibraryVersions:    NewLibraryVersions(),
		ByteStreamSettings: NewByteStreamSettings(),
		PRUDPV0Settings:    NewPRUDPV0Settings(),
		PRUDPV1Settings:    NewPRUDPV1Settings(),
	}
}
