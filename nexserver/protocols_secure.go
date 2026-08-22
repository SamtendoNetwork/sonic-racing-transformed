package nexserver

import (
	commonnattraversal "github.com/PretendoNetwork/nex-protocols-common-go/v2/nat-traversal"
	commonranking "github.com/PretendoNetwork/nex-protocols-common-go/v2/ranking"
	commonsecureconnection "github.com/PretendoNetwork/nex-protocols-common-go/v2/secure-connection"
	matchmakingprotocol "github.com/PretendoNetwork/nex-protocols-go/v2/match-making"
	matchmakingextprotocol "github.com/PretendoNetwork/nex-protocols-go/v2/match-making-ext"
	matchmakeextensionprotocol "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
	nattraversalprotocol "github.com/PretendoNetwork/nex-protocols-go/v2/nat-traversal"
	rankingprotocol "github.com/PretendoNetwork/nex-protocols-go/v2/ranking"
	secureconnectionprotocol "github.com/PretendoNetwork/nex-protocols-go/v2/secure-connection"
	"github.com/nex-server/sonic-transformed/globals"
	"github.com/nex-server/sonic-transformed/matchmaking"
)

// registerSecureProtocols wires every protocol the game can reach onto the
// secure endpoint.
//
// The protocol set is not guesswork: kinnay's RPX-derived library index shows
// every Sonic Transformed build links exactly three NEX libraries -
//
//	nex    3.0.1   core: Secure Connection (0xB), NAT Traversal (0x3)
//	nexmm  3.0.1   MatchMaking (0x15), MatchMakingExt (0x32), MatchmakeExtension (0x6D)
//	nexrk  3.0.1   Ranking (0x70)
//
// and notably does NOT link nexds (DataStore) or nexut (Utility). Registering
// those would be dead code at best; registering Utility would be actively
// harmful, since utility, storage-manager and subscriber all claim protocol ID
// 0x6E and RegisterServiceProtocol just appends another handler.
func registerSecureProtocols() {
	registerSecureConnection()
	registerNATTraversal()
	registerMatchmaking()
	registerRanking()
}

// registerSecureConnection wires Secure Connection (0xB).
//
// This is where a console publishes its station URLs - the addresses other
// consoles will use to reach it directly during a race. common-go's Register
// and RegisterEx already do the important part: they overwrite the client's
// claimed public address with the address the packet actually arrived from, so
// a console behind NAT advertises its real external address rather than its
// LAN address.
func registerSecureConnection() {
	protocol := secureconnectionprotocol.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(protocol)

	common := commonsecureconnection.NewCommonProtocol(protocol)

	// * RegisterEx refuses to run without a validator, exactly as LoginEx does.
	common.ValidateLoginData = validateLoginData

	// * SendReport carries the game's own telemetry/abuse reports. There is no
	// * report database here, so they are logged and acknowledged.
	common.CreateReportDBRecord = createReportRecord

	// * NEX 3.0.1 predates the switch to RegisterEx being mandatory, and a 2013
	// * title may still call the plain Register. Enabling it costs nothing:
	// * both paths do the same station URL reflection.
	common.EnableInsecureRegister()

	// * The three methods common-go does not implement.
	protocol.SetHandlerRequestConnectionData(requestConnectionData)
	protocol.SetHandlerTestConnectivity(testConnectivity)
	protocol.SetHandlerUpdateURLs(updateURLs)
}

// registerNATTraversal wires NAT Traversal (0x3).
//
// The server's role here is to relay probe requests between consoles so they
// can punch a hole through each other's NAT. common-go implements everything
// except the older non-Ext probe request, which a NEX 3.0.1 title may use.
func registerNATTraversal() {
	protocol := nattraversalprotocol.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(protocol)

	commonnattraversal.NewCommonProtocol(protocol)

	protocol.SetHandlerRequestProbeInitiation(requestProbeInitiation)
}

// registerMatchmaking wires all three matchmaking protocols to the shared
// in-memory store.
func registerMatchmaking() {
	matchMaking := matchmakingprotocol.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchMaking)

	matchMakingExt := matchmakingextprotocol.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchMakingExt)

	matchmakeExtension := matchmakeextensionprotocol.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchmakeExtension)

	protocol := matchmaking.NewProtocol(Matchmaking, globals.Settings.AllowPublicMatchmaking, globals.Logger)
	protocol.Register(globals.SecureEndpoint, matchMaking, matchMakingExt, matchmakeExtension)

	if !globals.Settings.AllowPublicMatchmaking {
		globals.Logger.Warning("Public matchmaking is disabled; only direct joins by gathering ID will work")
	}
}

// registerRanking wires Ranking (0x70) to the persistent leaderboard store.
//
// common-go's ranking protocol is entirely callback driven - no database, no
// manager - so the store plugs straight in.
func registerRanking() {
	protocol := rankingprotocol.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(protocol)

	common := commonranking.NewCommonProtocol(protocol)

	common.GetCommonData = Ranking.GetCommonData
	common.UploadCommonData = Ranking.UploadCommonData
	common.InsertRankingByPIDAndRankingScoreData = Ranking.InsertRankingByPIDAndRankingScoreData
	common.GetRankingsAndCountByCategoryAndRankingOrderParam = Ranking.GetRankingsAndCountByCategoryAndRankingOrderParam
	common.GetNearbyRankingsAndCountByCategoryAndRankingOrderParam = Ranking.GetNearbyRankingsAndCountByCategoryAndRankingOrderParam
	common.GetFriendsRankingsAndCountByCategoryAndRankingOrderParam = Ranking.GetFriendsRankingsAndCountByCategoryAndRankingOrderParam
	common.GetNearbyFriendsRankingsAndCountByCategoryAndRankingOrderParam = Ranking.GetNearbyFriendsRankingsAndCountByCategoryAndRankingOrderParam
	common.GetOwnRankingByCategoryAndRankingOrderParam = Ranking.GetOwnRankingByCategoryAndRankingOrderParam

	// * The methods common-go leaves unregistered.
	protocol.SetHandlerDeleteScore(deleteScore)
	protocol.SetHandlerDeleteAllScores(deleteAllScores)
	protocol.SetHandlerDeleteCommonData(deleteCommonData)
	protocol.SetHandlerGetRankingByPIDList(getRankingByPIDList)
	protocol.SetHandlerGetApproxOrder(getApproxOrder)
}
