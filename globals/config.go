package globals

import (
	"os"
	"strconv"
	"strings"
)













const (
	GameServerID = 0x10111F00
	TitleID      = 0x0005000010111F00

	DefaultAccessKey = "b26a3421"




	BuildName = "branch:origin/release/ngs/3.4.x.3 build:3_4_13_3_0"























	NEXVersionMajor = 3
	NEXVersionMinor = 0
	NEXVersionPatch = 1
)


type Config struct {

	AuthenticationServerPort int

	SecureServerPort int






	SecureServerHost string




	AccessKey string





	KerberosPassword string


	AccountGRPCHost   string
	AccountGRPCPort   int
	AccountGRPCAPIKey string


	DataDir string



	AllowPublicMatchmaking bool



	LogPackets bool





	StructureHeader string
}


var Settings Config


func LoadConfig() {
	Settings = Config{
		AuthenticationServerPort: envInt("SONIC_AUTHENTICATION_SERVER_PORT", 60000),
		SecureServerPort:         envInt("SONIC_SECURE_SERVER_PORT", 60001),
		SecureServerHost:         envString("SONIC_SECURE_SERVER_HOST", "127.0.0.1"),
		AccessKey:                envString("SONIC_ACCESS_KEY", DefaultAccessKey),
		KerberosPassword:         envString("SONIC_KERBEROS_PASSWORD", "sonic-transformed"),
		AccountGRPCHost:          envString("SONIC_ACCOUNT_GRPC_HOST", ""),
		AccountGRPCPort:          envInt("SONIC_ACCOUNT_GRPC_PORT", 0),
		AccountGRPCAPIKey:        envString("SONIC_ACCOUNT_GRPC_API_KEY", ""),
		DataDir:                  envString("SONIC_DATA_DIR", "data"),
		AllowPublicMatchmaking:   envBool("SONIC_ALLOW_PUBLIC_MATCHMAKING", true),
		LogPackets:               envBool("SONIC_LOG_PACKETS", false),
		StructureHeader:          strings.ToLower(envString("SONIC_STRUCTURE_HEADER", "auto")),
	}
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
