package globals

import (
	"context"
	"math"
	"net"
	"testing"
	"time"

	pb "github.com/PretendoNetwork/grpc/go/account/v2"
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	tg "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type testAccounts struct {
	pb.UnimplementedAccountServiceServer
	password string
	response *pb.ExchangeNEXTokenForUserDataResponse
	failure  error
}

func (s *testAccounts) GetNEXPassword(ctx context.Context, req *pb.GetNEXPasswordRequest) (*pb.GetNEXPasswordResponse, error) {
	if req.Pid != 100 {
		return nil, status.Error(codes.InvalidArgument, "unknown account")
	}
	if s.failure != nil {
		return nil, s.failure
	}
	return &pb.GetNEXPasswordResponse{Password: s.password}, nil
}
func (s *testAccounts) ExchangeNEXTokenForUserData(ctx context.Context, req *pb.ExchangeNEXTokenForUserDataRequest) (*pb.ExchangeNEXTokenForUserDataResponse, error) {
	if req.Token != "opaque+token/unchanged=" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}
	return s.response, s.failure
}
func setupAccounts(t *testing.T) *testAccounts {
	t.Helper()
	service := &testAccounts{password: "stored-password"}
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		keys := md.Get("x-api-key")
		if len(keys) != 1 || keys[0] != "test-api-key" {
			return nil, status.Error(codes.Unauthenticated, "API key required")
		}
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > accountRPCTimeout {
			return nil, status.Error(codes.InvalidArgument, "deadline required")
		}
		return handler(ctx, req)
	}))
	pb.RegisterAccountServiceServer(server, service)
	go server.Serve(listener)
	conn, err := grpc.NewClient("passthrough:///accounts", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	oldClient, oldMetadata := GRPCAccountClient, GRPCAccountCommonMetadata
	GRPCAccountClient = pb.NewAccountServiceClient(conn)
	GRPCAccountCommonMetadata = metadata.Pairs("x-api-key", "test-api-key")
	t.Cleanup(func() {
		conn.Close()
		server.Stop()
		listener.Close()
		GRPCAccountClient = oldClient
		GRPCAccountCommonMetadata = oldMetadata
	})
	return service
}
func validTokenResponse() *pb.ExchangeNEXTokenForUserDataResponse {
	return &pb.ExchangeNEXTokenForUserDataResponse{
		NexAccount: &pb.NEXAccount{Pid: 100},
		TokenInfo:  &pb.TokenInfo{Pid: 100, TokenType: pb.TokenType_TOKEN_TYPE_NEX, SystemType: pb.SystemType_SYSTEM_TYPE_WUP, ExpireTime: timestamppb.New(time.Now().Add(time.Hour))},
	}
}
func loginHolder(authenticationInfo bool, token string) types.DataHolder {
	holder := types.NewDataHolder()
	if authenticationInfo {
		data := tg.NewAuthenticationInfo()
		data.Token = types.NewString(token)
		holder.Object = data
	} else {
		data := tg.NewNintendoLoginData()
		data.Token = types.NewString(token)
		holder.Object = data
	}
	return holder
}
func TestAccountsPasswordLookup(t *testing.T) {
	service := setupAccounts(t)
	password, code := PasswordFromPID(types.NewPID(100))
	if password != "stored-password" || code != 0 {
		t.Fatalf("password lookup failed: %q %d", password, code)
	}
	for _, pid := range []uint64{0, 101, math.MaxUint32 + 101} {
		if password, code := PasswordFromPID(types.NewPID(pid)); password != "" || code == 0 {
			t.Fatalf("accepted invalid PID %d", pid)
		}
	}
	service.password = ""
	if _, code := PasswordFromPID(types.NewPID(100)); code == 0 {
		t.Fatal("accepted empty password")
	}
	service.failure = status.Error(codes.Unavailable, "offline")
	if _, code := PasswordFromPID(types.NewPID(100)); code == 0 {
		t.Fatal("accepted unavailable accounts")
	}
}
func TestAccountsTokenValidation(t *testing.T) {
	service := setupAccounts(t)
	tests := []struct {
		name   string
		mutate func(*pb.ExchangeNEXTokenForUserDataResponse)
		want   uint32
	}{
		{"valid", func(*pb.ExchangeNEXTokenForUserDataResponse) {}, 0},
		{"PID mismatch", func(r *pb.ExchangeNEXTokenForUserDataResponse) { r.TokenInfo.Pid = 101 }, nex.ResultCodes.Authentication.PrincipalIDUnmatched},
		{"account mismatch", func(r *pb.ExchangeNEXTokenForUserDataResponse) { r.NexAccount.Pid = 101 }, nex.ResultCodes.Authentication.PrincipalIDUnmatched},
		{"expired", func(r *pb.ExchangeNEXTokenForUserDataResponse) {
			r.TokenInfo.ExpireTime = timestamppb.New(time.Now().Add(-time.Hour))
		}, nex.ResultCodes.Authentication.TokenExpired},
		{"missing expiry", func(r *pb.ExchangeNEXTokenForUserDataResponse) { r.TokenInfo.ExpireTime = nil }, nex.ResultCodes.Authentication.TokenExpired},
		{"malformed expiry", func(r *pb.ExchangeNEXTokenForUserDataResponse) {
			r.TokenInfo.ExpireTime = &timestamppb.Timestamp{Nanos: -1}
		}, nex.ResultCodes.Authentication.TokenExpired},
		{"wrong type", func(r *pb.ExchangeNEXTokenForUserDataResponse) {
			r.TokenInfo.TokenType = pb.TokenType_TOKEN_TYPE_OAUTH_ACCESS
		}, nex.ResultCodes.Authentication.ValidationFailed},
		{"wrong system", func(r *pb.ExchangeNEXTokenForUserDataResponse) {
			r.TokenInfo.SystemType = pb.SystemType_SYSTEM_TYPE_CTR
		}, nex.ResultCodes.Authentication.ValidationFailed},
		{"disabled token", func(r *pb.ExchangeNEXTokenForUserDataResponse) { r.TokenInfo.AccessLevel = -1 }, nex.ResultCodes.RendezVous.AccountDisabled},
		{"disabled account", func(r *pb.ExchangeNEXTokenForUserDataResponse) { r.NexAccount.AccessLevel = -1 }, nex.ResultCodes.RendezVous.AccountDisabled},
		{"missing account", func(r *pb.ExchangeNEXTokenForUserDataResponse) { r.NexAccount = nil }, nex.ResultCodes.Authentication.ValidationFailed},
		{"missing info", func(r *pb.ExchangeNEXTokenForUserDataResponse) { r.TokenInfo = nil }, nex.ResultCodes.Authentication.ValidationFailed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service.response = validTokenResponse()
			tc.mutate(service.response)
			for _, authInfo := range []bool{false, true} {
				err := ValidateLoginData(types.NewPID(100), loginHolder(authInfo, "opaque+token/unchanged="))
				if tc.want == 0 {
					if err != nil {
						t.Fatal(err)
					}
				} else if err == nil || err.ResultCode != nex.NewError(tc.want, "").ResultCode {
					t.Fatalf("got %v, want %d", err, tc.want)
				}
			}
		})
	}
	service.response = validTokenResponse()
	for _, holder := range []types.DataHolder{types.NewDataHolder(), loginHolder(false, ""), loginHolder(false, "invalid")} {
		if err := ValidateLoginData(types.NewPID(100), holder); err == nil {
			t.Fatal("accepted invalid login")
		}
	}
	service.failure = status.Error(codes.Unavailable, "offline")
	if err := ValidateLoginData(types.NewPID(100), loginHolder(false, "opaque+token/unchanged=")); err == nil {
		t.Fatal("accepted offline accounts")
	}
	service.failure = nil
	GRPCAccountCommonMetadata = metadata.Pairs("x-api-key", "wrong")
	if err := ValidateLoginData(types.NewPID(100), loginHolder(false, "opaque+token/unchanged=")); err == nil {
		t.Fatal("accepted wrong API key")
	}
	if _, code := PasswordFromPID(types.NewPID(100)); code == 0 {
		t.Fatal("accepted wrong API key for password")
	}
}

func TestAccountCallbacksUseAccountsAndHaveNoGuestFallback(t *testing.T) {
	service := setupAccounts(t)
	oldAuth, oldSecure := AuthenticationServerAccount, SecureServerAccount
	AuthenticationServerAccount = nex.NewAccount(types.NewPID(1), "Quazal Authentication", "internal-key", false)
	SecureServerAccount = nex.NewAccount(types.NewPID(2), "Quazal Rendez-Vous", "internal-key", false)
	t.Cleanup(func() { AuthenticationServerAccount, SecureServerAccount = oldAuth, oldSecure })
	for _, principal := range []*nex.Account{AuthenticationServerAccount, SecureServerAccount} {
		byPID, err := AccountDetailsByPID(principal.PID)
		if err != nil || byPID != principal {
			t.Fatal("internal principal lookup failed")
		}
		byName, err := AccountDetailsByUsername(principal.Username)
		if err != nil || byName != principal {
			t.Fatal("internal principal name lookup failed")
		}
	}

	byPID, err := AccountDetailsByPID(types.NewPID(100))
	if err != nil || byPID.Password != service.password {
		t.Fatal("PID lookup did not use accounts")
	}
	byName, err := AccountDetailsByUsername("100")
	if err != nil || byName.Password != service.password {
		t.Fatal("username lookup did not use accounts")
	}
	for _, username := range []string{"guest", "unknown", "-1", "0", "4294967296"} {
		if _, err := AccountDetailsByUsername(username); err == nil {
			t.Fatalf("accepted %q", username)
		}
	}
	service.failure = status.Error(codes.Unavailable, "offline")
	if _, err := AccountDetailsByPID(types.NewPID(100)); err == nil {
		t.Fatal("guest fallback during outage")
	}
	GRPCAccountClient = nil
	if _, code := PasswordFromPID(types.NewPID(100)); code == 0 {
		t.Fatal("password fallback without accounts")
	}
	if err := ValidateLoginData(types.NewPID(100), loginHolder(false, "opaque+token/unchanged=")); err == nil {
		t.Fatal("token accepted without accounts")
	}
}
