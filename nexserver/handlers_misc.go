package nexserver

import (
	"encoding/hex"
	"net"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/PretendoNetwork/nex-go/v2/types"
	nattraversal "github.com/PretendoNetwork/nex-protocols-go/v2/nat-traversal"
	ticketgranting "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting"
	"github.com/nex-server/sonic-transformed/globals"
)




func observedAddress(connection *nex.PRUDPConnection) (string, uint16) {
	switch address := connection.Address().(type) {
	case *net.UDPAddr:
		return address.IP.String(), uint16(address.Port)
	case *net.TCPAddr:
		return address.IP.String(), uint16(address.Port)
	default:
		return "", 0
	}
}














func requestProbeInitiation(err error, packet nex.PacketInterface, callID uint32, urlTargetList types.List[types.StationURL]) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, err.Error())
	}

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)
	server := endpoint.Server


	stationToProbe := types.NewStationURL("")
	for _, stationURL := range connection.StationURLs {
		if stationURL.IsPublic() {
			stationToProbe = stationURL.Copy().(types.StationURL)
			break
		}
	}

	requestStream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())
	stationToProbe.WriteTo(requestStream)

	request := nex.NewRMCRequest(endpoint)
	request.ProtocolID = nattraversal.ProtocolID
	request.MethodID = nattraversal.MethodInitiateProbe
	request.CallID = 0xFFFF0000 + callID
	request.Parameters = requestStream.Bytes()

	requestBytes := request.Bytes()

	for _, target := range urlTargetList {
		connectionID, ok := target.RVConnectionID()
		if !ok {
			continue
		}

		peer := endpoint.FindConnectionByID(connectionID)
		if peer == nil {
			globals.Logger.Warningf("RequestProbeInitiation: no connection with RVCID %d", connectionID)
			continue
		}

		var probe nex.PRUDPPacketInterface

		switch peer.DefaultPRUDPVersion {
		case 0:
			probe, _ = nex.NewPRUDPPacketV0(server, peer, nil)
		case 1:
			probe, _ = nex.NewPRUDPPacketV1(server, peer, nil)
		case 2:
			probe, _ = nex.NewPRUDPPacketLite(server, peer, nil)
		default:
			globals.Logger.Errorf("RequestProbeInitiation: unsupported PRUDP version %d", peer.DefaultPRUDPVersion)
			continue
		}

		probe.SetType(constants.DataPacket)
		probe.AddFlag(constants.PacketFlagNeedsAck)
		probe.AddFlag(constants.PacketFlagReliable)
		probe.SetSourceVirtualPortStreamType(peer.StreamType)
		probe.SetSourceVirtualPortStreamID(endpoint.StreamID)
		probe.SetDestinationVirtualPortStreamType(peer.StreamType)
		probe.SetDestinationVirtualPortStreamID(peer.StreamID)
		probe.SetPayload(requestBytes)

		server.Send(probe)
	}

	response := nex.NewRMCSuccess(endpoint, nil)
	response.ProtocolID = nattraversal.ProtocolID
	response.MethodID = nattraversal.MethodRequestProbeInitiation
	response.CallID = callID

	return response, nil
}








func getPID(err error, packet nex.PacketInterface, callID uint32, strUserName types.String) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, err.Error())
	}

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)

	account, nexError := globals.AccountDetailsByUsername(string(strUserName))
	if nexError != nil {
		return nil, nexError
	}

	stream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())
	account.PID.WriteTo(stream)

	response := nex.NewRMCSuccess(endpoint, stream.Bytes())
	response.ProtocolID = ticketgranting.ProtocolID
	response.MethodID = ticketgranting.MethodGetPID
	response.CallID = callID

	return response, nil
}




func getName(err error, packet nex.PacketInterface, callID uint32, id types.PID) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, err.Error())
	}

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)

	account, nexError := globals.AccountDetailsByPID(id)
	if nexError != nil {
		return nil, nexError
	}

	stream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())
	types.NewString(account.Username).WriteTo(stream)

	response := nex.NewRMCSuccess(endpoint, stream.Bytes())
	response.ProtocolID = ticketgranting.ProtocolID
	response.MethodID = ticketgranting.MethodGetName
	response.CallID = callID

	return response, nil
}











func traceHandler(tag string) func(packet nex.PacketInterface) {
	return func(packet nex.PacketInterface) {
		if !globals.Settings.LogPackets {
			return
		}

		request := packet.RMCMessage()
		connection := packet.Sender().(*nex.PRUDPConnection)

		globals.Logger.Infof(
			"[%s] pid=%d protocol=%#x method=%#x call=%d params=%d bytes %s",
			tag,
			uint64(connection.PID()),
			request.ProtocolID,
			request.MethodID,
			request.CallID,
			len(request.Parameters),
			hex.EncodeToString(request.Parameters),
		)
	}
}
