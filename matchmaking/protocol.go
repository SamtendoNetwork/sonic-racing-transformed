package matchmaking

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	matchmakingprotocol "github.com/PretendoNetwork/nex-protocols-go/v2/match-making"
	matchmakingextprotocol "github.com/PretendoNetwork/nex-protocols-go/v2/match-making-ext"
	matchmakingtypes "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
	matchmakeextensionprotocol "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
	notificationsconstants "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/constants"
	"github.com/PretendoNetwork/plogger-go"
)










type Protocol struct {
	Store *Store




	AllowPublicMatchmaking bool

	Logger *plogger.Logger
}


func NewProtocol(store *Store, allowPublicMatchmaking bool, logger *plogger.Logger) *Protocol {
	return &Protocol{
		Store:                  store,
		AllowPublicMatchmaking: allowPublicMatchmaking,
		Logger:                 logger,
	}
}



func (p *Protocol) Register(
	endpoint *nex.PRUDPEndPoint,
	matchMaking *matchmakingprotocol.Protocol,
	matchMakingExt *matchmakingextprotocol.Protocol,
	matchmakeExtension *matchmakeextensionprotocol.Protocol,
) {
	p.registerMatchmakeExtension(matchmakeExtension)
	p.registerMatchMaking(matchMaking)
	p.registerMatchMakingExt(matchMakingExt)




	endpoint.OnConnectionEnded(func(connection *nex.PRUDPConnection) {
		p.Store.RemoveConnection(uint64(connection.PID()), notificationsconstants.ParticipationEventsDisconnect)
	})
}






func connectionOf(packet nex.PacketInterface) (*nex.PRUDPConnection, *nex.PRUDPEndPoint) {
	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)

	return connection, endpoint
}


func newStream(endpoint *nex.PRUDPEndPoint) *nex.ByteStreamOut {
	return nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())
}


func respond(endpoint *nex.PRUDPEndPoint, protocolID uint16, methodID uint32, callID uint32, body []byte) *nex.RMCMessage {
	response := nex.NewRMCSuccess(endpoint, body)
	response.ProtocolID = protocolID
	response.MethodID = methodID
	response.CallID = callID

	return response
}


func respondEmpty(endpoint *nex.PRUDPEndPoint, protocolID uint16, methodID uint32, callID uint32) *nex.RMCMessage {
	return respond(endpoint, protocolID, methodID, callID, nil)
}


func invalidArgument(err error) *nex.Error {
	if err != nil {
		return nex.NewError(nex.ResultCodes.Core.InvalidArgument, err.Error())
	}

	return nex.NewError(nex.ResultCodes.Core.InvalidArgument, "Invalid argument")
}


func holderFor(session matchmakingtypes.MatchmakeSession) matchmakingtypes.GatheringHolder {
	holder := matchmakingtypes.NewGatheringHolder()
	holder.Object = session.Copy().(matchmakingtypes.GatheringInterface)

	return holder
}




func publicCopy(session matchmakingtypes.MatchmakeSession) matchmakingtypes.MatchmakeSession {
	copied := session.Copy().(matchmakingtypes.MatchmakeSession)
	copied.SessionKey = types.NewBuffer(make([]byte, 0))
	copied.UserPassword = types.NewString("")

	return copied
}


func sessionFromHolder(holder matchmakingtypes.GatheringHolder) (matchmakingtypes.MatchmakeSession, *nex.Error) {
	if holder.Object == nil {
		return matchmakingtypes.MatchmakeSession{}, invalidArgument(nil)
	}

	if !holder.Object.GatheringObjectID().Equals(types.NewString("MatchmakeSession")) {
		return matchmakingtypes.MatchmakeSession{}, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "Gathering is not a MatchmakeSession")
	}

	session, ok := holder.Object.(matchmakingtypes.MatchmakeSession)
	if !ok {
		return matchmakingtypes.MatchmakeSession{}, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "Gathering is not a MatchmakeSession")
	}

	return session, nil
}
