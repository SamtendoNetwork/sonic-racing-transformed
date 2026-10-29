package nex

import (
	"net"

	"github.com/PretendoNetwork/nex-go/v2/types"
)


type HPPClient struct {
	address  *net.TCPAddr
	endpoint *HPPServer
	pid      types.PID
}


func (c *HPPClient) Endpoint() EndpointInterface {
	return c.endpoint
}


func (c *HPPClient) Address() net.Addr {
	return c.address
}


func (c *HPPClient) PID() types.PID {
	return c.pid
}


func (c *HPPClient) SetPID(pid types.PID) {
	c.pid = pid
}


func NewHPPClient(address *net.TCPAddr, server *HPPServer) *HPPClient {
	return &HPPClient{
		address:  address,
		endpoint: server,
	}
}
