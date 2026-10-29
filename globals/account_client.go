package globals

import (
	"net"
	"strconv"

	pb "github.com/PretendoNetwork/grpc/go/account/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var GRPCAccountClient pb.AccountServiceClient
var GRPCAccountCommonMetadata metadata.MD




func ConnectAccounts() (*grpc.ClientConn, error) {
	address := net.JoinHostPort(Settings.AccountGRPCHost, strconv.Itoa(Settings.AccountGRPCPort))
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	GRPCAccountClient = pb.NewAccountServiceClient(connection)
	GRPCAccountCommonMetadata = metadata.Pairs("x-api-key", Settings.AccountGRPCAPIKey)
	return connection, nil
}
