package common_globals

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "github.com/PretendoNetwork/grpc/go/account/v2"
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	account_management_types "github.com/PretendoNetwork/nex-protocols-go/v2/account-management/types"
	ticket_granting_types "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)




func ConnectToAccountGRPC(host string, port uint16, apiKey string) {
	if GRPCAccountClientConnection != nil {
		return
	}

	var err error

	GRPCAccountClientConnection, err = grpc.NewClient(fmt.Sprintf("%s:%d", host, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		Logger.Criticalf("Failed to connect to account gRPC server: %v", err)
		os.Exit(0)
	}

	GRPCAccountClient = pb.NewAccountServiceClient(GRPCAccountClientConnection)
	GRPCAccountCommonMetadata = metadata.Pairs(
		"X-API-Key", apiKey,
	)
}


func ValidatePNLoginData(pid types.PID, loginData types.DataHolder, gameServerIDs []string) *nex.Error {
	var tokenBase64 string

	loginDataType := loginData.Object.DataObjectID().(types.String)

	switch loginDataType {
	case "NintendoLoginData":
		nintendoCreateAccountData := loginData.Object.Copy().(ticket_granting_types.NintendoLoginData)

		tokenBase64 = string(nintendoCreateAccountData.Token)
	case "AccountExtraInfo":
		accountExtraInfo := loginData.Object.Copy().(account_management_types.AccountExtraInfo)

		tokenBase64 = string(accountExtraInfo.NEXToken)
	case "AuthenticationInfo":
		authenticationInfo := loginData.Object.Copy().(ticket_granting_types.AuthenticationInfo)

		tokenBase64 = string(authenticationInfo.Token)
	default:
		Logger.Errorf("Invalid loginData data type %s!", loginDataType)
		return nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, fmt.Sprintf("Invalid loginData data type %s!", loginDataType))
	}

	ctx := metadata.NewOutgoingContext(context.Background(), GRPCAccountCommonMetadata)

	response, err := GRPCAccountClient.ExchangeNEXTokenForUserData(ctx, &pb.ExchangeNEXTokenForUserDataRequest{
		GameServerIds: gameServerIDs,
		Token:         tokenBase64,
	})
	if err != nil {


		return nex.NewError(nex.ResultCodes.Authentication.TokenExpired, err.Error())
	}




	if response.TokenInfo.TokenType != 3 {
		return nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Invalid token")
	}

	if response.TokenInfo.SystemType != 1 && response.TokenInfo.SystemType != 2 {
		return nex.NewError(nex.ResultCodes.Authentication.ValidationFailed, "Invalid token")
	}



	if response.TokenInfo.ExpireTime != nil && response.TokenInfo.ExpireTime.AsTime().Before(time.Now()) {
		return nex.NewError(nex.ResultCodes.Authentication.TokenExpired, "Token expired")
	}

	if types.NewPID(uint64(response.NexAccount.Pid)) != pid {
		return nex.NewError(nex.ResultCodes.Authentication.PrincipalIDUnmatched, fmt.Sprintf("Account %d expected, got %d", pid, response.NexAccount.Pid))
	}

	if response.NexAccount.AccessLevel < 0 {
		return nex.NewError(nex.ResultCodes.RendezVous.AccountDisabled, fmt.Sprintf("Account %d is banned", response.NexAccount.Pid))
	}

	return nil
}
