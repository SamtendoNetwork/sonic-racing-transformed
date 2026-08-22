package globals

import (
	"os"
	"strconv"
	"strings"
)

// * Sonic & All-Stars Racing Transformed (Wii U) server identity.
// *
// * These values come from the NEX game server list at
// * https://kinnay.github.io/view.html?page=nexwiiu ("Sonic Transformed"):
// *
// *   Game server ID  10111F00
// *   Title ID        0005000010111F00
// *   Access key      b26a3421
// *   Build           branch:origin/release/ngs/3.4.x.3 build:3_4_13_3_0
// *
// * The build string is echoed back to the console verbatim in the
// * TicketGranting::Login(Ex) response, so it must match the real one.
const (
	GameServerID = 0x10111F00
	TitleID      = 0x0005000010111F00

	DefaultAccessKey = "b26a3421"

	// * BuildName is echoed back to the console verbatim as strReturnMsg in the
	// * TicketGranting::Login(Ex) response. It is free text, not a version gate,
	// * so we return the authentic server build string for this title.
	BuildName = "branch:origin/release/ngs/3.4.x.3 build:3_4_13_3_0"

	// * The NEX *library* version linked into the game, which is what actually
	// * determines the wire format. Do not confuse it with the 3.4.13.3 build
	// * string above: that is the version of the server Nintendo/SEGA ran, and
	// * ~19 unrelated third-party Wii U titles all reported the same one.
	// *
	// * kinnay's RPX-derived library index (https://kinnay.github.io/data/wiiu.json)
	// * lists every Sonic Transformed build as linking:
	// *
	// *   nex 3.0.1, nexrk 3.0.1, nexmm 3.0.1
	// *
	// * and notably NOT nexds (DataStore) or nexut (Utility), which is why this
	// * server implements neither.
	// *
	// * At 3.0.1 the version-gated wire decisions are:
	// *   - structures have NO version/length header      (added at 3.5.0)
	// *   - MatchmakeSession has SessionKey               (added at 3.0.0)
	// *     but NOT ProgressScore                         (added at 3.4.0)
	// *   - MatchmakeSessionSearchCriteria has SelectionMethod (3.0.0)
	// *     but NOT VacantParticipants                    (added at 3.4.0)
	// *   - NotificationEvent stops after StrParam; no Param3 (added at 3.4.0)
	// *   - RVConnectionData stays at StructureVersion 0, so no Time field
	// *   - AuthenticationInfo has TokenType and ServerVersion (3.0.0)
	NEXVersionMajor = 3
	NEXVersionMinor = 0
	NEXVersionPatch = 1
)

// Config holds every runtime-tunable setting, all sourced from the environment.
type Config struct {
	// AuthenticationServerPort is the UDP port the ticket granting server binds.
	AuthenticationServerPort int
	// SecureServerPort is the UDP port the secure (game) server binds.
	SecureServerPort int

	// SecureServerHost is the address handed to the console inside the
	// TicketGranting::Login response so it knows where to find the secure
	// server. This MUST be an address the console can actually reach - a LAN
	// IP for LAN play, or a public IP/hostname for internet play. "localhost"
	// only works if the client runs on this same machine.
	SecureServerHost string

	// AccessKey is the game's 8 character hex access key. Every PRUDP packet
	// signature is derived from it, so a wrong value means the console's
	// packets are silently rejected.
	AccessKey string

	// KerberosPassword is the password for the two special server accounts
	// ("Quazal Authentication" / "Quazal Rendez-Vous"). It never leaves the
	// server, so its value is arbitrary, but it must be identical between the
	// authentication and secure servers.
	KerberosPassword string

	// AccountsFile optionally points at a JSON file mapping NEX PIDs to their
	// Kerberos passwords. See globals/accounts.go.
	AccountsFile string

	// AccountSecret seeds the deterministic password derivation used when a PID
	// is not listed in AccountsFile. An account server paired with this game
	// server can use the same secret to hand consoles a matching password.
	AccountSecret string

	// AllowUnknownAccounts lets any PID authenticate using the derived
	// password. With it disabled only PIDs in AccountsFile may connect.
	AllowUnknownAccounts bool

	// DataDir is where persistent state (leaderboards, common data) is stored.
	DataDir string

	// AllowPublicMatchmaking exposes sessions to BrowseMatchmakeSession and the
	// automatic matchmaker. Disabling it restricts play to friend/invite flows.
	AllowPublicMatchmaking bool

	// LogPackets prints every inbound RMC protocol/method pair. Extremely
	// noisy, but the fastest way to learn what the game actually calls.
	LogPackets bool

	// StructureHeader controls whether NEX structures carry a version/length
	// header: "auto" (detect from the console's first login), "on", or "off".
	// See nexserver/structure_header.go for why this cannot simply be derived
	// from the NEX version.
	StructureHeader string
}

// Settings is the process-wide configuration, populated by LoadConfig.
var Settings Config

// LoadConfig reads configuration from the environment, applying defaults.
func LoadConfig() {
	Settings = Config{
		AuthenticationServerPort: envInt("SONIC_AUTHENTICATION_SERVER_PORT", 60000),
		SecureServerPort:         envInt("SONIC_SECURE_SERVER_PORT", 60001),
		SecureServerHost:         envString("SONIC_SECURE_SERVER_HOST", "127.0.0.1"),
		AccessKey:                envString("SONIC_ACCESS_KEY", DefaultAccessKey),
		KerberosPassword:         envString("SONIC_KERBEROS_PASSWORD", "sonic-transformed"),
		AccountsFile:             envString("SONIC_ACCOUNTS_FILE", ""),
		AccountSecret:            envString("SONIC_ACCOUNT_SECRET", ""),
		AllowUnknownAccounts:     envBool("SONIC_ALLOW_UNKNOWN_ACCOUNTS", true),
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
