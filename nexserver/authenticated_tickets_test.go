package nexserver

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	ticketgranting "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting"
	tgtypes "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting/types"
)

func TestTicketRequestsRequireSuccessfulMatchingLogin(t *testing.T) {
	endpoint := nex.NewPRUDPEndPoint(1)
	protocol := ticketgranting.NewProtocol()
	connection := nex.NewPRUDPConnection(nil)
	packet, _ := nex.NewPRUDPPacketV0(nil, connection, nil)
	otherPacket, _ := nex.NewPRUDPPacketV0(nil, nex.NewPRUDPConnection(nil), nil)
	pid, securePID := types.NewPID(1800000001), types.NewPID(2)
	data := types.NewDataHolder()
	loginData := tgtypes.NewNintendoLoginData()
	loginData.Token = types.NewString("opaque+token/unchanged=")
	data.Object = loginData
	var loginResult uint32
	var loginFailure, tokenFailure *nex.Error
	tickets, validations := 0, 0
	protocol.SetHandlerLoginEx(func(error, nex.PacketInterface, uint32, types.String, types.DataHolder) (*nex.RMCMessage, *nex.Error) {
		return &nex.RMCMessage{IsSuccess: true, Parameters: binary.LittleEndian.AppendUint32(nil, loginResult)}, loginFailure
	})
	protocol.SetHandlerRequestTicket(func(error, nex.PacketInterface, uint32, types.PID, types.PID) (*nex.RMCMessage, *nex.Error) {
		tickets++
		return &nex.RMCMessage{IsSuccess: true}, nil
	})
	guardTicketRequests(protocol, endpoint, securePID, func(source types.PID, holder types.DataHolder) *nex.Error {
		validations++
		if source != pid || holder.Object.(tgtypes.NintendoLoginData).Token != loginData.Token {
			t.Fatal("ticket validation lost the login identity or token")
		}
		return tokenFailure
	})
	login := func() { protocol.LoginEx(nil, packet, 1, "1800000001", data) }
	reject := func(requestPacket nex.PacketInterface, source, target types.PID) {
		t.Helper()
		before := tickets
		if _, err := protocol.RequestTicket(nil, requestPacket, 2, source, target); err == nil || tickets != before {
			t.Fatal("unauthorized request reached ticket generation")
		}
	}
	reject(packet, pid, securePID)
	login()
	if response, err := protocol.RequestTicket(nil, packet, 2, pid, securePID); err != nil || !response.IsSuccess || tickets != 1 || validations != 1 {
		t.Fatal("valid LoginEx followed by RequestTicket failed")
	}
	reject(otherPacket, pid, securePID)
	reject(packet, types.NewPID(1800000002), securePID)
	reject(packet, pid, types.NewPID(1))
	connection.SessionID++
	reject(packet, pid, securePID)
	login()
	if _, err := protocol.RequestTicket(errors.New("bad packet"), packet, 2, pid, securePID); err == nil {
		t.Fatal("accepted malformed request")
	}
	// A failed login is an RMC success carrying an error QResult. It must also
	// invalidate an earlier successful login on the same connection.
	loginResult = nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "rejected").ResultCode
	login()
	reject(packet, pid, securePID)
	loginResult = 0
	loginFailure = nex.NewError(nex.ResultCodes.Core.InvalidArgument, "decode failure")
	login()
	reject(packet, pid, securePID)
	loginFailure = nil
	login()
	tokenFailure = nex.NewError(nex.ResultCodes.Authentication.TokenExpired, "expired")
	reject(packet, pid, securePID)
	tokenFailure = nil
	reject(packet, pid, securePID) // Revalidation failure clears the login.
}
