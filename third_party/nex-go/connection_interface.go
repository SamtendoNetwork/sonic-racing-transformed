
package nex

import (
	"net"

	"github.com/PretendoNetwork/nex-go/v2/types"
)


type ConnectionInterface interface {
	Endpoint() EndpointInterface
	Address() net.Addr
	PID() types.PID
	SetPID(pid types.PID)
}
