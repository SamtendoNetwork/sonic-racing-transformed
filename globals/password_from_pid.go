package globals

import (
	"context"
	"math"
	"time"

	pb "github.com/PretendoNetwork/grpc/go/account/v2"
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	"google.golang.org/grpc/metadata"
)

const accountRPCTimeout = 5 * time.Second

func accountContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(metadata.NewOutgoingContext(context.Background(), GRPCAccountCommonMetadata), accountRPCTimeout)
}

func PasswordFromPID(pid types.PID) (string, uint32) {
	if uint64(pid) > math.MaxUint32 || pid == 0 || GRPCAccountClient == nil {
		return "", nex.ResultCodes.RendezVous.InvalidUsername
	}
	ctx, cancel := accountContext()
	defer cancel()
	response, err := GRPCAccountClient.GetNEXPassword(ctx, &pb.GetNEXPasswordRequest{Pid: uint32(pid)})
	if err != nil || response == nil || response.Password == "" {
		return "", nex.ResultCodes.RendezVous.InvalidUsername
	}
	return response.Password, 0
}
