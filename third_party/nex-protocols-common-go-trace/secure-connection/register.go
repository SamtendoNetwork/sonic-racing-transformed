package secureconnection

import (
	"fmt"
	"net"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/PretendoNetwork/nex-go/v2/types"
	secure_connection "github.com/PretendoNetwork/nex-protocols-go/v2/secure-connection"

	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
)

func (commonProtocol *CommonProtocol) register(err error, packet nex.PacketInterface, callID uint32, vecMyURLs types.List[types.StationURL]) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		common_globals.Logger.Error(err.Error())
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, err.Error())
	}

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint()

	var retval types.QResult
	pidConnectionID := types.NewUInt32(0)
	urlPublic := types.NewString("")

	if !commonProtocol.allowInsecureRegisterMethod {
		common_globals.Logger.Error("SecureConnection::Register blocked")
		retval = types.NewQResultError(nex.ResultCodes.Authentication.ValidationFailed)
	} else {

		var localStation *types.StationURL
		var publicStation *types.StationURL

		for _, stationURL := range vecMyURLs {
			natf, ok := stationURL.NATFiltering()
			if !ok {
				continue
			}

			natm, ok := stationURL.NATMapping()
			if !ok {
				continue
			}


			if localStation == nil && !stationURL.IsPublic() {
				localStation = &stationURL
			}


			if localStation == nil && natf == constants.UnknownNATFiltering && natm == constants.UnknownNATMapping {
				localStation = &stationURL
			}

			if publicStation == nil && stationURL.IsPublic() {
				publicStation = &stationURL
			}
		}

		if localStation == nil {
			common_globals.Logger.Error("Failed to find local station")
			return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
		}

		if publicStation == nil {
			publicStation = localStation

			var address string
			var port uint16


			switch clientAddress := connection.Address().(type) {
			case *net.UDPAddr:
				address = clientAddress.IP.String()
				port = uint16(clientAddress.Port)
			case *net.TCPAddr:
				address = clientAddress.IP.String()
				port = uint16(clientAddress.Port)
			}

			publicStation.SetAddress(address)
			publicStation.SetPortNumber(port)
			publicStation.SetNATFiltering(constants.UnknownNATFiltering)
			publicStation.SetNATMapping(constants.UnknownNATMapping)
			publicStation.SetType(uint8(constants.StationURLFlagPublic) | uint8(constants.StationURLFlagBehindNAT))
		}

		localStation.SetPrincipalID(connection.PID())
		publicStation.SetPrincipalID(connection.PID())

		localStation.SetRVConnectionID(connection.ID)
		publicStation.SetRVConnectionID(connection.ID)





		address, _ := publicStation.Address()
		port, _ := publicStation.PortNumber()
		natf, _ := publicStation.NATFiltering()
		natm, _ := publicStation.NATMapping()
		sid, _ := publicStation.StreamID()
		pmp := 0
		if publicStation.IsNATPMPSupported() {
			pmp = 1
		}
		upnp := 0
		if publicStation.IsUPnPSupported() {
			upnp = 1
		}
		responsePublicURL := fmt.Sprintf(
			"prudp:/address=%s;port=%d;natf=%d;natm=%d;pmp=%d;sid=%d;upnp=%d;PID=%d;RVCID=%d;type=3",
			address, port, natf, natm, pmp, sid, upnp, connection.PID(), connection.ID,
		)

		connection.StationURLs = append(connection.StationURLs, *localStation)
		connection.StationURLs = append(connection.StationURLs, *publicStation)

		retval = types.NewQResultSuccess(nex.ResultCodes.Core.Unknown)
		pidConnectionID = types.NewUInt32(connection.ID)
		urlPublic = types.NewString(responsePublicURL)
	}

	rmcResponseStream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())

	retval.WriteTo(rmcResponseStream)
	pidConnectionID.WriteTo(rmcResponseStream)
	urlPublic.WriteTo(rmcResponseStream)

	rmcResponseBody := rmcResponseStream.Bytes()
	common_globals.Logger.Infof(
		"[REGISTER-RESPONSE] pid=%d cid=%d url=%q body=%x",
		connection.PID(), connection.ID, string(urlPublic), rmcResponseBody)

	rmcResponse := nex.NewRMCSuccess(endpoint, rmcResponseBody)
	rmcResponse.ProtocolID = secure_connection.ProtocolID
	rmcResponse.MethodID = secure_connection.MethodRegister
	rmcResponse.CallID = callID

	if commonProtocol.OnAfterRegister != nil {
		go commonProtocol.OnAfterRegister(packet, vecMyURLs)
	}

	return rmcResponse, nil
}
