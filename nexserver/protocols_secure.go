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















func registerSecureProtocols() {
	registerSecureConnection()
	registerNATTraversal()
	registerMatchmaking()
	registerRanking()
}









func registerSecureConnection() {
	protocol := secureconnectionprotocol.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(protocol)

	common := commonsecureconnection.NewCommonProtocol(protocol)


	common.ValidateLoginData = globals.ValidateLoginData



	common.CreateReportDBRecord = createReportRecord




	common.EnableInsecureRegister()


	protocol.SetHandlerRequestConnectionData(requestConnectionData)
	protocol.SetHandlerTestConnectivity(testConnectivity)
	protocol.SetHandlerUpdateURLs(updateURLs)
}






func registerNATTraversal() {
	protocol := nattraversalprotocol.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(protocol)

	commonnattraversal.NewCommonProtocol(protocol)

	protocol.SetHandlerRequestProbeInitiation(requestProbeInitiation)
}



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


	protocol.SetHandlerDeleteScore(deleteScore)
	protocol.SetHandlerDeleteAllScores(deleteAllScores)
	protocol.SetHandlerDeleteCommonData(deleteCommonData)
	protocol.SetHandlerGetRankingByPIDList(getRankingByPIDList)
	protocol.SetHandlerGetApproxOrder(getApproxOrder)
}
