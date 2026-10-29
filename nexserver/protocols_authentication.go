package nexserver

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/PretendoNetwork/nex-go/v2/types"
	commonticketgranting "github.com/PretendoNetwork/nex-protocols-common-go/v2/ticket-granting"
	ticketgranting "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting"
	"github.com/nex-server/sonic-transformed/globals"
)



func registerAuthenticationProtocols() {
	ticketGrantingProtocol := ticketgranting.NewProtocol()
	globals.AuthenticationEndpoint.RegisterServiceProtocol(ticketGrantingProtocol)

	commonTicketGranting := commonticketgranting.NewCommonProtocol(ticketGrantingProtocol)









	secureStationURL := types.NewStationURL("")
	secureStationURL.SetURLType(constants.StationURLPRUDPS)
	secureStationURL.SetAddress(globals.Settings.SecureServerHost)
	secureStationURL.SetPortNumber(uint16(globals.Settings.SecureServerPort))
	secureStationURL.SetConnectionID(1)
	secureStationURL.SetPrincipalID(globals.SecureServerAccount.PID)
	secureStationURL.SetStreamID(1)
	secureStationURL.SetStreamType(constants.StreamTypeRVSecure)
	secureStationURL.SetType(uint8(constants.StationURLFlagPublic))

	commonTicketGranting.SecureStationURL = secureStationURL
	commonTicketGranting.SecureServerAccount = globals.SecureServerAccount




	commonTicketGranting.BuildName = types.NewString(globals.BuildName)




	commonTicketGranting.SessionKeyLength = 32


	commonTicketGranting.ValidateLoginData = globals.ValidateLoginData
	commonTicketGranting.DisableInsecureLogin()

	ticketGrantingProtocol.SetHandlerRequestTicket(rejectTokenlessTicket)


	ticketGrantingProtocol.SetHandlerGetPID(getPID)
	ticketGrantingProtocol.SetHandlerGetName(getName)
}


func rejectTokenlessTicket(err error, packet nex.PacketInterface, callID uint32, source types.PID, target types.PID) (*nex.RMCMessage, *nex.Error) {
	return nil, nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Use LoginEx with an accounts-issued NEX token")
}
