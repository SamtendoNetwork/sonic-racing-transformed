package nexserver

import (
	"fmt"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/nex-server/sonic-transformed/globals"
)




const (



	sessionKeyLength = 32




	kerberosTicketVersion = 0
)















func StartAuthenticationServer() {
	globals.AuthenticationServer = nex.NewPRUDPServer()




	globals.AuthenticationServer.ByteStreamSettings.UseStructureHeader = false

	globals.AuthenticationServer.LibraryVersions.SetDefault(nex.NewLibraryVersion(
		globals.NEXVersionMajor,
		globals.NEXVersionMinor,
		globals.NEXVersionPatch,
	))
	globals.AuthenticationServer.AccessKey = globals.Settings.AccessKey
	globals.AuthenticationServer.PRUDPV1Settings.LegacyConnectionSignature = true
	globals.AuthenticationServer.SessionKeyLength = sessionKeyLength
	globals.AuthenticationServer.KerberosTicketVersion = kerberosTicketVersion



	globals.AuthenticationEndpoint = nex.NewPRUDPEndPoint(1)
	globals.AuthenticationEndpoint.ServerAccount = globals.AuthenticationServerAccount
	globals.AuthenticationEndpoint.AccountDetailsByPID = globals.AccountDetailsByPID
	globals.AuthenticationEndpoint.AccountDetailsByUsername = globals.AccountDetailsByUsername

	globals.AuthenticationServer.BindPRUDPEndPoint(globals.AuthenticationEndpoint)




	globals.AuthenticationEndpoint.OnData(detectStructureHeader)
	globals.AuthenticationEndpoint.OnData(traceHandler("AUTH"))

	globals.AuthenticationEndpoint.OnError(func(err *nex.Error) {
		globals.Logger.Errorf("[AUTH] %v", err)
	})

	applyStructureHeaderMode()

	registerAuthenticationProtocols()

	globals.Logger.Successf(
		"Authentication server listening on UDP :%d (access key %s, NEX %d.%d.%d)",
		globals.Settings.AuthenticationServerPort,
		globals.Settings.AccessKey,
		globals.NEXVersionMajor, globals.NEXVersionMinor, globals.NEXVersionPatch,
	)
	globals.Logger.Infof(
		"Consoles will be redirected to the secure server at %s",
		fmt.Sprintf("%s:%d", globals.Settings.SecureServerHost, globals.Settings.SecureServerPort),
	)

	globals.AuthenticationServer.Listen(globals.Settings.AuthenticationServerPort)
}
