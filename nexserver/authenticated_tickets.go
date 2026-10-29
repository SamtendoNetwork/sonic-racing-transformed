package nexserver

import (
	"encoding/binary"
	"strconv"
	"sync"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	ticketgranting "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting"
)

// Sonic requests another ticket after LoginEx. Permit that only for the same
// user and PRUDP session, and revalidate the token before issuing the ticket.
// The common protocol's OnAfterLoginEx hook cannot establish this state: it is
// asynchronous and runs even when the login returns an error result.
func guardTicketRequests(protocol *ticketgranting.Protocol, endpoint *nex.PRUDPEndPoint, securePID types.PID, validate func(types.PID, types.DataHolder) *nex.Error) {
	type login struct {
		pid       types.PID
		sessionID uint8
		data      types.DataHolder
	}
	var mutex sync.Mutex
	logins := make(map[*nex.PRUDPConnection]login)
	clear := func(connection *nex.PRUDPConnection) {
		mutex.Lock()
		delete(logins, connection)
		mutex.Unlock()
	}
	endpoint.OnConnectionEnded(clear)

	loginEx := protocol.LoginEx
	requestTicket := protocol.RequestTicket
	protocol.SetHandlerLoginEx(func(err error, packet nex.PacketInterface, callID uint32, username types.String, data types.DataHolder) (*nex.RMCMessage, *nex.Error) {
		connection := packet.Sender().(*nex.PRUDPConnection)
		clear(connection)
		response, loginError := loginEx(err, packet, callID, username, data)
		// LoginEx reports authentication failures in its first QResult parameter,
		// even when the surrounding RMC message itself is successful.
		if loginError == nil && response != nil && response.IsSuccess && len(response.Parameters) >= 4 && binary.LittleEndian.Uint32(response.Parameters[:4])&0x80000000 == 0 {
			pid, parseError := strconv.ParseUint(string(username), 10, 32)
			if parseError == nil && pid != 0 {
				mutex.Lock()
				logins[connection] = login{types.NewPID(pid), connection.SessionID, data.Copy().(types.DataHolder)}
				mutex.Unlock()
			}
		}
		return response, loginError
	})
	protocol.SetHandlerRequestTicket(func(err error, packet nex.PacketInterface, callID uint32, source, target types.PID) (*nex.RMCMessage, *nex.Error) {
		if err != nil {
			return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "Invalid ticket request")
		}
		connection := packet.Sender().(*nex.PRUDPConnection)
		mutex.Lock()
		authenticated, ok := logins[connection]
		mutex.Unlock()
		if !ok || authenticated.sessionID != connection.SessionID || source != authenticated.pid || target != securePID {
			return nil, nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Ticket request requires a matching successful LoginEx")
		}
		if validationError := validate(source, authenticated.data); validationError != nil {
			clear(connection)
			return nil, validationError
		}
		return requestTicket(nil, packet, callID, source, target)
	})
}
