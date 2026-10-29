package nex

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/PretendoNetwork/nex-go/v2/types"
)


type HPPServer struct {
	server                   *http.Server
	accessKey                string
	libraryVersions          *LibraryVersions
	dataHandlers             []func(packet PacketInterface)
	errorEventHandlers       []func(err *Error)
	byteStreamSettings       *ByteStreamSettings
	AccountDetailsByPID      func(pid types.PID) (*Account, *Error)
	AccountDetailsByUsername func(username string) (*Account, *Error)
	useVerboseRMC            bool
}


func (s *HPPServer) RegisterServiceProtocol(protocol ServiceProtocol) {
	protocol.SetEndpoint(s)
	s.OnData(protocol.HandlePacket)
}


func (s *HPPServer) OnData(handler func(packet PacketInterface)) {
	s.dataHandlers = append(s.dataHandlers, handler)
}


func (s *HPPServer) EmitError(err *Error) {
	for _, handler := range s.errorEventHandlers {
		go handler(err)
	}
}

func (s *HPPServer) handleRequest(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pidValue := req.Header.Get("pid")
	if pidValue == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}


	token := req.Header.Get("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accessKeySignature := req.Header.Get("signature1")
	if accessKeySignature == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	passwordSignature := req.Header.Get("signature2")
	if passwordSignature == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pid, err := strconv.Atoi(pidValue)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rmcRequestString := req.FormValue("file")

	rmcRequestBytes := []byte(rmcRequestString)

	tcpAddr, err := net.ResolveTCPAddr("tcp", req.RemoteAddr)
	if err != nil {

		logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	client := NewHPPClient(tcpAddr, s)
	client.SetPID(types.NewPID(uint64(pid)))

	hppPacket, err := NewHPPPacket(client, rmcRequestBytes)
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = hppPacket.validateAccessKeySignature(accessKeySignature)
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = hppPacket.validatePasswordSignature(passwordSignature)
	if err != nil {
		logger.Error(err.Error())

		rmcMessage := hppPacket.RMCMessage()


		errorResponse := NewRMCError(s, ResultCodes.PythonCore.ValidationError)
		errorResponse.CallID = rmcMessage.CallID
		errorResponse.IsHPP = true

		_, err = w.Write(errorResponse.Bytes())
		if err != nil {
			logger.Error(err.Error())
		}

		return
	}

	for _, dataHandler := range s.dataHandlers {
		go dataHandler(hppPacket)
	}

	<-hppPacket.processed

	if len(hppPacket.payload) > 0 {
		_, err = w.Write(hppPacket.payload)
		if err != nil {
			logger.Error(err.Error())
		}
	}
}


func (s *HPPServer) Listen(port int) {
	s.server.Addr = fmt.Sprintf(":%d", port)

	err := s.server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}


func (s *HPPServer) ListenSecure(port int, certFile, keyFile string) {
	s.server.Addr = fmt.Sprintf(":%d", port)

	err := s.server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		panic(err)
	}
}


func (s *HPPServer) Send(packet PacketInterface) {
	if packet, ok := packet.(*HPPPacket); ok {
		packet.message.IsHPP = true
		packet.payload = packet.message.Bytes()

		packet.processed <- true
	}
}


func (s *HPPServer) LibraryVersions() *LibraryVersions {
	return s.libraryVersions
}


func (s *HPPServer) AccessKey() string {
	return s.accessKey
}


func (s *HPPServer) SetAccessKey(accessKey string) {
	s.accessKey = accessKey
}


func (s *HPPServer) ByteStreamSettings() *ByteStreamSettings {
	return s.byteStreamSettings
}


func (s *HPPServer) SetByteStreamSettings(byteStreamSettings *ByteStreamSettings) {
	s.byteStreamSettings = byteStreamSettings
}


func (s *HPPServer) UseVerboseRMC() bool {
	return s.useVerboseRMC
}


func (s *HPPServer) EnableVerboseRMC(enable bool) {
	s.useVerboseRMC = enable
}


func NewHPPServer() *HPPServer {
	s := &HPPServer{
		dataHandlers:       make([]func(packet PacketInterface), 0),
		errorEventHandlers: make([]func(err *Error), 0),
		libraryVersions:    NewLibraryVersions(),
		byteStreamSettings: NewByteStreamSettings(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hpp/", s.handleRequest)

	httpServer := &http.Server{
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS11,
		},
	}

	s.server = httpServer

	return s
}
