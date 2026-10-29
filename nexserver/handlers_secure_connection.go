package nexserver

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	secureconnection "github.com/PretendoNetwork/nex-protocols-go/v2/secure-connection"
	"github.com/nex-server/sonic-transformed/globals"
)














func requestConnectionData(err error, packet nex.PacketInterface, callID uint32, cidTarget types.UInt32, pidTarget types.PID) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, err.Error())
	}

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)

	target := endpoint.FindConnectionByID(uint32(cidTarget))
	if target == nil {
		target = endpoint.FindConnectionByPID(uint64(pidTarget))
	}

	stream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())

	if target == nil {
		globals.Logger.Warningf("RequestConnectionData: no connection for CID %d / PID %d", uint32(cidTarget), uint64(pidTarget))

		types.NewBool(false).WriteTo(stream)
		stream.WriteUInt32LE(0)
	} else {
		types.NewBool(true).WriteTo(stream)


		stream.WriteUInt32LE(uint32(len(target.StationURLs)))

		for _, stationURL := range target.StationURLs {
			stationURL.WriteTo(stream)
			types.NewUInt32(target.ID).WriteTo(stream)
		}
	}

	response := nex.NewRMCSuccess(endpoint, stream.Bytes())
	response.ProtocolID = secureconnection.ProtocolID
	response.MethodID = secureconnection.MethodRequestConnectionData
	response.CallID = callID

	return response, nil
}





func testConnectivity(err error, packet nex.PacketInterface, callID uint32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, err.Error())
	}

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)

	response := nex.NewRMCSuccess(endpoint, nil)
	response.ProtocolID = secureconnection.ProtocolID
	response.MethodID = secureconnection.MethodTestConnectivity
	response.CallID = callID

	return response, nil
}









func updateURLs(err error, packet nex.PacketInterface, callID uint32, vecMyURLs types.List[types.StationURL]) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, err.Error())
	}

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)

	updated := make(types.List[types.StationURL], 0, len(vecMyURLs))

	for _, stationURL := range vecMyURLs {
		station := stationURL.Copy().(types.StationURL)

		station.SetPrincipalID(connection.PID())
		station.SetRVConnectionID(connection.ID)

		if station.IsPublic() {
			address, port := observedAddress(connection)
			if address != "" {
				station.SetAddress(address)
				station.SetPortNumber(port)
			}
		}

		updated = append(updated, station)
	}

	connection.StationURLs = updated

	if globals.Settings.LogPackets {
		globals.Logger.Infof("PID %d updated %d station URL(s)", uint64(connection.PID()), len(updated))
	}

	response := nex.NewRMCSuccess(endpoint, nil)
	response.ProtocolID = secureconnection.ProtocolID
	response.MethodID = secureconnection.MethodUpdateURLs
	response.CallID = callID

	return response, nil
}






func createReportRecord(pid types.PID, reportID types.UInt32, reportData types.QBuffer) error {
	globals.Logger.Infof("Report from PID %d: id=%d, %d bytes", uint64(pid), uint32(reportID), len(reportData))

	return nil
}
