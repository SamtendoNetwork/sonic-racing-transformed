







package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"





	_ "github.com/PretendoNetwork/nex-protocols-go/v2"

	"github.com/PretendoNetwork/plogger-go"
	"github.com/joho/godotenv"
	"github.com/nex-server/sonic-transformed/globals"
	"github.com/nex-server/sonic-transformed/nexserver"
	"github.com/nex-server/sonic-transformed/ranking"
)

func main() {
	globals.Logger = plogger.NewLogger()

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		globals.Logger.Warningf("Could not read .env: %v", err)
	}

	globals.LoadConfig()

	if err := validateConfig(); err != nil {
		globals.Logger.Criticalf("%v", err)
		os.Exit(1)
	}

	globals.InitAccounts()

	accountConnection, err := globals.ConnectAccounts()
	if err != nil {
		globals.Logger.Criticalf("%v", err)
		os.Exit(1)
	}

	defer accountConnection.Close()

	rankingStore, err := ranking.NewStore(globals.Settings.DataDir)
	if err != nil {
		globals.Logger.Criticalf("Failed to open the ranking store: %v", err)
		os.Exit(1)
	}

	nexserver.Ranking = rankingStore

	banner()



	go nexserver.StartAuthenticationServer()
	go nexserver.StartSecureServer()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	<-shutdown

	globals.Logger.Info("Shutting down")
}



func validateConfig() error {
	settings := globals.Settings

	if settings.AuthenticationServerPort < 1 || settings.AuthenticationServerPort > 65535 {
		return fmt.Errorf("SONIC_AUTHENTICATION_SERVER_PORT %d is not a valid port", settings.AuthenticationServerPort)
	}

	if settings.SecureServerPort < 1 || settings.SecureServerPort > 65535 {
		return fmt.Errorf("SONIC_SECURE_SERVER_PORT %d is not a valid port", settings.SecureServerPort)
	}

	if settings.AuthenticationServerPort == settings.SecureServerPort {
		return fmt.Errorf("the authentication and secure servers cannot share port %d", settings.AuthenticationServerPort)
	}

	if len(settings.AccessKey) != 8 {
		return fmt.Errorf("SONIC_ACCESS_KEY %q is not 8 characters; the correct key for this game is %s", settings.AccessKey, globals.DefaultAccessKey)
	}

	if settings.SecureServerHost == "" {
		return fmt.Errorf("SONIC_SECURE_SERVER_HOST must be set to an address the console can reach")
	}

	if settings.AccountGRPCHost == "" {
		return fmt.Errorf("SONIC_ACCOUNT_GRPC_HOST must be set to the accounts service host")
	}
	if settings.AccountGRPCPort < 1 || settings.AccountGRPCPort > 65535 {
		return fmt.Errorf("SONIC_ACCOUNT_GRPC_PORT must be a valid accounts gRPC port")
	}
	if settings.AccountGRPCAPIKey == "" {
		return fmt.Errorf("SONIC_ACCOUNT_GRPC_API_KEY must match accounts' PN_ACT_CONFIG_GRPC_MASTER_API_KEY_ACCOUNT")
	}

	return nil
}

func banner() {
	globals.Logger.Success("Sonic & All-Stars Racing Transformed - NEX server")
	globals.Logger.Infof("  Game server ID   %08X", globals.GameServerID)
	globals.Logger.Infof("  Access key       %s", globals.Settings.AccessKey)
	globals.Logger.Infof("  NEX version      %d.%d.%d", globals.NEXVersionMajor, globals.NEXVersionMinor, globals.NEXVersionPatch)
	globals.Logger.Infof("  Build string     %s", globals.BuildName)
	globals.Logger.Infof("  Data directory   %s", globals.Settings.DataDir)
}
