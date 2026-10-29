package globals

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	"github.com/PretendoNetwork/plogger-go"
)


var Logger *plogger.Logger










var (
	AuthenticationServer   *nex.PRUDPServer
	AuthenticationEndpoint *nex.PRUDPEndPoint

	SecureServer   *nex.PRUDPServer
	SecureEndpoint *nex.PRUDPEndPoint
)




var (
	AuthenticationServerAccount *nex.Account
	SecureServerAccount         *nex.Account
)


func InitAccounts() {
	AuthenticationServerAccount = nex.NewAccount(types.NewPID(1), "Quazal Authentication", Settings.KerberosPassword, false)
	SecureServerAccount = nex.NewAccount(types.NewPID(2), "Quazal Rendez-Vous", Settings.KerberosPassword, false)
}
