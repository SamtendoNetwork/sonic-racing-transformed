package nexserver

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/nex-server/sonic-transformed/globals"
	"github.com/nex-server/sonic-transformed/matchmaking"
	"github.com/nex-server/sonic-transformed/ranking"
)



var Matchmaking *matchmaking.Store


var Ranking *ranking.Store








func StartSecureServer() {
	globals.SecureServer = nex.NewPRUDPServer()
















	globals.SecureServer.ByteStreamSettings.UseStructureHeader = false
	globals.SecureServer.LibraryVersions.SetDefault(nex.NewLibraryVersion(
		globals.NEXVersionMajor,
		globals.NEXVersionMinor,
		globals.NEXVersionPatch,
	))
	globals.SecureServer.AccessKey = globals.Settings.AccessKey
	globals.SecureServer.PRUDPV1Settings.LegacyConnectionSignature = true
	globals.SecureServer.SessionKeyLength = sessionKeyLength
	globals.SecureServer.KerberosTicketVersion = kerberosTicketVersion

	globals.SecureEndpoint = nex.NewPRUDPEndPoint(1)




	globals.SecureEndpoint.IsSecureEndPoint = true

	globals.SecureEndpoint.ServerAccount = globals.SecureServerAccount
	globals.SecureEndpoint.AccountDetailsByPID = globals.AccountDetailsByPID
	globals.SecureEndpoint.AccountDetailsByUsername = globals.AccountDetailsByUsername

	globals.SecureServer.BindPRUDPEndPoint(globals.SecureEndpoint)




	globals.SecureEndpoint.OnData(detectStructureHeader)
	globals.SecureEndpoint.OnData(traceHandler("SECURE"))

	globals.SecureEndpoint.OnError(func(err *nex.Error) {
		globals.Logger.Errorf("[SECURE] %v", err)
	})

	globals.SecureEndpoint.OnConnectionEnded(func(connection *nex.PRUDPConnection) {
		globals.Logger.Infof("PID %d disconnected", uint64(connection.PID()))
	})

	Matchmaking = matchmaking.NewStore(globals.SecureEndpoint)
	Matchmaking.GetUserFriendPIDs = globals.GetUserFriendPIDs
	Ranking.GetUserFriendPIDs = globals.GetUserFriendPIDs

	registerSecureProtocols()

	globals.Logger.Successf("Secure server listening on UDP :%d", globals.Settings.SecureServerPort)

	globals.SecureServer.Listen(globals.Settings.SecureServerPort)
}
