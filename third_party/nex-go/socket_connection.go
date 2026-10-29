package nex

import (
	"net"

	"github.com/lxzan/gws"
)



type SocketConnection struct {
	Server              *PRUDPServer
	Address             net.Addr
	WebSocketConnection *gws.Conn
}


func NewSocketConnection(server *PRUDPServer, address net.Addr, webSocketConnection *gws.Conn) *SocketConnection {
	return &SocketConnection{
		Server:              server,
		Address:             address,
		WebSocketConnection: webSocketConnection,
	}
}
