package globals

import (
	"math"
	"time"

	pb "github.com/PretendoNetwork/grpc/go/account/v2"
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	ticket_granting_types "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting/types"
)

func ValidateLoginData(pid types.PID, loginData types.DataHolder) *nex.Error {
	if pid == 0 || uint64(pid) > math.MaxUint32 {
		return nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Invalid login PID")
	}


	var token string
	switch data := loginData.Object.(type) {
	case ticket_granting_types.NintendoLoginData:
		token = string(data.Token)
	case ticket_granting_types.AuthenticationInfo:
		token = string(data.Token)
	default:
		return nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Unsupported login data")
	}
	if token == "" || GRPCAccountClient == nil {
		return nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Missing token or accounts service")
	}
	ctx, cancel := accountContext()
	defer cancel()
	response, err := GRPCAccountClient.ExchangeNEXTokenForUserData(ctx, &pb.ExchangeNEXTokenForUserDataRequest{Token: token})
	if err != nil || response == nil || response.TokenInfo == nil || response.NexAccount == nil {
		return nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Accounts rejected the NEX token")
	}
	info := response.TokenInfo
	if info.TokenType != pb.TokenType_TOKEN_TYPE_NEX || info.SystemType != pb.SystemType_SYSTEM_TYPE_WUP {
		return nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Invalid token type or system")
	}
	if info.ExpireTime == nil || info.ExpireTime.CheckValid() != nil || !info.ExpireTime.AsTime().After(time.Now()) {
		return nex.NewError(nex.ResultCodes.Authentication.TokenExpired, "Token expired")
	}
	if info.Pid != uint64(pid) || uint64(response.NexAccount.Pid) != uint64(pid) {
		return nex.NewError(nex.ResultCodes.Authentication.PrincipalIDUnmatched, "Token PID does not match login PID")
	}
	if info.AccessLevel < 0 || response.NexAccount.AccessLevel < 0 {
		return nex.NewError(nex.ResultCodes.RendezVous.AccountDisabled, "Account is disabled")
	}
	return nil
}
