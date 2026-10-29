package nex

import (
	"fmt"
	"net/http"

	"github.com/lxzan/gws"
)

type wsEventHandler struct {
	prudpServer *PRUDPServer
}

func (wseh *wsEventHandler) OnOpen(socket *gws.Conn) {}

func (wseh *wsEventHandler) OnClose(wsConn *gws.Conn, _ error) {

	wseh.prudpServer.Endpoints.Each(func(streamid uint8, pep *PRUDPEndPoint) bool {
		connections := make([]*PRUDPConnection, 0)

		pep.Connections.Each(func(discriminator string, pc *PRUDPConnection) bool {
			if pc.Socket.Address == wsConn.RemoteAddr() {
				connections = append(connections, pc)
			}
			return false
		})





		for _, connection := range connections {
			pep.CleanupConnection(connection)
		}
		return false
	})
}

func (wseh *wsEventHandler) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.WritePong(nil)
}

func (wseh *wsEventHandler) OnPong(socket *gws.Conn, payload []byte) {}

func (wseh *wsEventHandler) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()




	packetData := append([]byte(nil), message.Bytes()...)
	err := wseh.prudpServer.handleSocketMessage(packetData, socket.RemoteAddr(), socket)
	if err != nil {
		logger.Error(err.Error())
	}
}


type WebSocketServer struct {
	mux         *http.ServeMux
	upgrader    *gws.Upgrader
	prudpServer *PRUDPServer
}

func (ws *WebSocketServer) init() {
	ws.upgrader = gws.NewUpgrader(&wsEventHandler{
		prudpServer: ws.prudpServer,
	}, &gws.ServerOption{
		ParallelEnabled: true,
		Recovery:        gws.Recovery,
		ReadBufferSize:  64000,
		WriteBufferSize: 64000,
	})

	ws.mux = http.NewServeMux()
	ws.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		socket, err := ws.upgrader.Upgrade(w, r)
		if err != nil {
			return
		}

		go func() {
			socket.ReadLoop()
		}()
	})
}

func (ws *WebSocketServer) listen(port int) {
	ws.init()

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), ws.mux)
	if err != nil {
		panic(err)
	}
}

func (ws *WebSocketServer) listenSecure(port int, certFile, keyFile string) {
	ws.init()

	err := http.ListenAndServeTLS(fmt.Sprintf(":%d", port), certFile, keyFile, ws.mux)
	if err != nil {
		panic(err)
	}
}
