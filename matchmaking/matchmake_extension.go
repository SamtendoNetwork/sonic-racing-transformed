package matchmaking

import (
	"unicode/utf8"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	commonglobals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	matchmakingtypes "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
	matchmakeextension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
	notificationsconstants "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/constants"
	notificationstypes "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/types"
)


const maxMessageLength = 256

func (p *Protocol) registerMatchmakeExtension(protocol *matchmakeextension.Protocol) {
	protocol.SetHandlerCloseParticipation(p.closeParticipation)
	protocol.SetHandlerOpenParticipation(p.openParticipation)
	protocol.SetHandlerAutoMatchmakePostpone(p.autoMatchmakePostpone)
	protocol.SetHandlerAutoMatchmakeWithSearchCriteriaPostpone(p.autoMatchmakeWithSearchCriteriaPostpone)
	protocol.SetHandlerBrowseMatchmakeSession(p.browseMatchmakeSession)
	protocol.SetHandlerBrowseMatchmakeSessionWithHostURLs(p.browseMatchmakeSessionWithHostURLs)
	protocol.SetHandlerCreateMatchmakeSession(p.createMatchmakeSession)
	protocol.SetHandlerJoinMatchmakeSession(p.joinMatchmakeSession)
	protocol.SetHandlerJoinMatchmakeSessionEx(p.joinMatchmakeSessionEx)
	protocol.SetHandlerModifyCurrentGameAttribute(p.modifyCurrentGameAttribute)
	protocol.SetHandlerUpdateApplicationBuffer(p.updateApplicationBuffer)
	protocol.SetHandlerUpdateMatchmakeSessionAttribute(p.updateMatchmakeSessionAttribute)
	protocol.SetHandlerUpdateMatchmakeSession(p.updateMatchmakeSession)
	protocol.SetHandlerUpdateProgressScore(p.updateProgressScore)
	protocol.SetHandlerUpdateNotificationData(p.updateNotificationData)
	protocol.SetHandlerGetFriendNotificationData(p.getFriendNotificationData)
	protocol.SetHandlerGetlstFriendNotificationData(p.getLstFriendNotificationData)
	protocol.SetHandlerGetPlayingSession(p.getPlayingSession)
	protocol.SetHandlerGetSimplePlayingSession(p.getSimplePlayingSession)
	protocol.SetHandlerFindMatchmakeSessionByGatheringID(p.findMatchmakeSessionByGatheringID)
	protocol.SetHandlerFindMatchmakeSessionBySingleGatheringID(p.findMatchmakeSessionBySingleGatheringID)
	protocol.SetHandlerFindMatchmakeSessionByOwner(p.findMatchmakeSessionByOwner)
}









func (p *Protocol) createMatchmakeSession(err error, packet nex.PacketInterface, callID uint32, anyGathering matchmakingtypes.GatheringHolder, message types.String, participationCount types.UInt16) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	if len(message) > maxMessageLength {
		return nil, invalidArgument(nil)
	}

	requested, nexError := sessionFromHolder(anyGathering)
	if nexError != nil {
		return nil, nexError
	}

	if !commonglobals.CheckValidMatchmakeSession(requested) {
		return nil, invalidArgument(nil)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()




	p.Store.RemoveConnectionLocked(uint64(connection.PID()), notificationsconstants.ParticipationEventsDisconnect)

	session := p.Store.CreateLocked(connection, requested)

	if _, nexError = p.Store.JoinLocked(session, connection, uint16(participationCount), string(message)); nexError != nil {
		delete(p.Store.sessions, uint32(session.MatchmakeSession.ID))
		p.Store.Unlock()
		return nil, nexError
	}

	created := session.MatchmakeSession.Copy().(matchmakingtypes.MatchmakeSession)

	p.Store.Unlock()

	p.Logger.Infof("Gathering %d created by PID %d (max %d players)", uint32(created.ID), uint64(connection.PID()), uint16(created.MaximumParticipants))

	stream := newStream(endpoint)
	created.ID.WriteTo(stream)

	if endpoint.Server.LibraryVersions.MatchMaking.GreaterOrEqual("3.0.0") {
		created.SessionKey.WriteTo(stream)
	}

	return respond(endpoint, matchmakeextension.ProtocolID, matchmakeextension.MethodCreateMatchmakeSession, callID, stream.Bytes()), nil
}




func (p *Protocol) joinMatchmakeSession(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32, strMessage types.String) (*nex.RMCMessage, *nex.Error) {
	return p.join(err, packet, callID, gid, strMessage, 1, matchmakeextension.MethodJoinMatchmakeSession)
}



func (p *Protocol) joinMatchmakeSessionEx(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32, strMessage types.String, dontCareMyBlockList types.Bool, participationCount types.UInt16) (*nex.RMCMessage, *nex.Error) {
	return p.join(err, packet, callID, gid, strMessage, uint16(participationCount), matchmakeextension.MethodJoinMatchmakeSessionEx)
}

func (p *Protocol) join(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32, strMessage types.String, participationCount uint16, methodID uint32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	if len(strMessage) > maxMessageLength {
		return nil, invalidArgument(nil)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()

	session, nexError := p.Store.GetLocked(uint32(gid))
	if nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	if bool(session.MatchmakeSession.UserPasswordEnabled) || bool(session.MatchmakeSession.SystemPasswordEnabled) {
		p.Store.Unlock()
		return nil, nex.NewError(nex.ResultCodes.RendezVous.PermissionDenied, "Gathering is password protected")
	}

	if nexError = p.Store.CanJoin(connection.PID(), session); nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	if _, nexError = p.Store.JoinLocked(session, connection, participationCount, string(strMessage)); nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	sessionKey := session.MatchmakeSession.SessionKey.Copy().(types.Buffer)

	p.Store.Unlock()

	p.Logger.Infof("PID %d joined gathering %d", uint64(connection.PID()), uint32(gid))

	stream := newStream(endpoint)

	if endpoint.Server.LibraryVersions.MatchMaking.GreaterOrEqual("3.0.0") {
		sessionKey.WriteTo(stream)
	}

	return respond(endpoint, matchmakeextension.ProtocolID, methodID, callID, stream.Bytes()), nil
}











func (p *Protocol) autoMatchmakePostpone(err error, packet nex.PacketInterface, callID uint32, anyGathering matchmakingtypes.GatheringHolder, message types.String) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	if len(message) > maxMessageLength {
		return nil, invalidArgument(nil)
	}

	requested, nexError := sessionFromHolder(anyGathering)
	if nexError != nil {
		return nil, nexError
	}



	return p.autoMatchmake(
		packet,
		callID,
		nil,
		requested,
		message,
		matchmakeextension.MethodAutoMatchmakePostpone,
	)
}




func (p *Protocol) autoMatchmakeWithSearchCriteriaPostpone(err error, packet nex.PacketInterface, callID uint32, lstSearchCriteria types.List[matchmakingtypes.MatchmakeSessionSearchCriteria], anyGathering matchmakingtypes.GatheringHolder, strMessage types.String) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	if len(strMessage) > maxMessageLength {
		return nil, invalidArgument(nil)
	}

	requested, nexError := sessionFromHolder(anyGathering)
	if nexError != nil {
		return nil, nexError
	}

	return p.autoMatchmake(
		packet,
		callID,
		lstSearchCriteria,
		requested,
		strMessage,
		matchmakeextension.MethodAutoMatchmakeWithSearchCriteriaPostpone,
	)
}

func (p *Protocol) autoMatchmake(
	packet nex.PacketInterface,
	callID uint32,
	searchCriterias []matchmakingtypes.MatchmakeSessionSearchCriteria,
	requested matchmakingtypes.MatchmakeSession,
	message types.String,
	methodID uint32,
) (*nex.RMCMessage, *nex.Error) {
	if !commonglobals.CheckValidMatchmakeSession(requested) {
		return nil, invalidArgument(nil)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()



	p.Store.RemoveConnectionLocked(uint64(connection.PID()), notificationsconstants.ParticipationEventsDisconnect)

	var session *Session

	if len(searchCriterias) == 0 {

		if candidate := p.Store.FindSession(connection, requested, p.AllowPublicMatchmaking); candidate != nil {
			if p.Store.CanJoin(connection.PID(), candidate) == nil {
				session = candidate
			}
		}
	} else {


		resultRange := types.NewResultRange()
		resultRange.Offset = types.NewUInt32(0)
		resultRange.Length = types.NewUInt32(1)

		candidates := p.Store.Search(connection, searchCriterias, resultRange, &requested, SearchOptions{
			AutoMatchmake:   true,
			ExcludeOwnerPID: uint64(connection.PID()),
			AllowPublic:     p.AllowPublicMatchmaking,
		})

		for _, candidate := range candidates {
			if p.Store.CanJoin(connection.PID(), candidate) == nil {
				session = candidate
				break
			}
		}
	}

	created := false
	if session == nil {

		session = p.Store.CreateLocked(connection, requested)
		created = true
	}

	participants, nexError := p.Store.JoinLocked(session, connection, 1, string(message))
	if nexError != nil {
		if created {
			delete(p.Store.sessions, uint32(session.MatchmakeSession.ID))
		}

		p.Store.Unlock()
		return nil, nexError
	}

	result := session.MatchmakeSession.Copy().(matchmakingtypes.MatchmakeSession)
	result.ParticipationCount = types.NewUInt32(participants)

	p.Store.Unlock()

	if created {
		p.Logger.Infof("PID %d started a new gathering %d via auto-matchmake", uint64(connection.PID()), uint32(result.ID))
	} else {
		p.Logger.Infof("PID %d auto-matched into gathering %d (%d players)", uint64(connection.PID()), uint32(result.ID), participants)
	}

	stream := newStream(endpoint)
	holderFor(result).WriteTo(stream)

	return respond(endpoint, matchmakeextension.ProtocolID, methodID, callID, stream.Bytes()), nil
}








func (p *Protocol) browseMatchmakeSession(err error, packet nex.PacketInterface, callID uint32, searchCriteria matchmakingtypes.MatchmakeSessionSearchCriteria, resultRange types.ResultRange) (*nex.RMCMessage, *nex.Error) {
	return p.browse(err, packet, callID, searchCriteria, resultRange, matchmakeextension.MethodBrowseMatchmakeSession)
}





func (p *Protocol) browseMatchmakeSessionWithHostURLs(err error, packet nex.PacketInterface, callID uint32, searchCriteria matchmakingtypes.MatchmakeSessionSearchCriteria, resultRange types.ResultRange) (*nex.RMCMessage, *nex.Error) {
	return p.browse(err, packet, callID, searchCriteria, resultRange, matchmakeextension.MethodBrowseMatchmakeSessionWithHostURLs)
}

func (p *Protocol) browse(err error, packet nex.PacketInterface, callID uint32, searchCriteria matchmakingtypes.MatchmakeSessionSearchCriteria, resultRange types.ResultRange, methodID uint32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.RLock()

	sessions := p.Store.Search(
		connection,
		[]matchmakingtypes.MatchmakeSessionSearchCriteria{searchCriteria},
		resultRange,
		nil,
		SearchOptions{AllowPublic: p.AllowPublicMatchmaking},
	)

	gatherings := types.NewList[matchmakingtypes.GatheringHolder]()
	for _, session := range sessions {
		gatherings = append(gatherings, holderFor(publicCopy(session.MatchmakeSession)))
	}

	p.Store.RUnlock()

	stream := newStream(endpoint)
	gatherings.WriteTo(stream)

	return respond(endpoint, matchmakeextension.ProtocolID, methodID, callID, stream.Bytes()), nil
}











func (p *Protocol) findMatchmakeSessionByGatheringID(err error, packet nex.PacketInterface, callID uint32, lstGID types.List[types.UInt32]) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()

	sessions := types.NewList[matchmakingtypes.MatchmakeSession]()
	for _, gid := range lstGID {
		if session, nexError := p.Store.GetLocked(uint32(gid)); nexError == nil {
			sessions = append(sessions, publicCopy(session.MatchmakeSession))
		}
	}

	p.Store.RUnlock()

	stream := newStream(endpoint)
	sessions.WriteTo(stream)

	return respond(endpoint, matchmakeextension.ProtocolID, matchmakeextension.MethodFindMatchmakeSessionByGatheringID, callID, stream.Bytes()), nil
}





func (p *Protocol) findMatchmakeSessionBySingleGatheringID(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()
	session, nexError := p.Store.GetLocked(uint32(gid))
	if nexError != nil {
		p.Store.RUnlock()
		return nil, nexError
	}

	found := publicCopy(session.MatchmakeSession)
	p.Store.RUnlock()

	stream := newStream(endpoint)
	found.WriteTo(stream)

	return respond(endpoint, matchmakeextension.ProtocolID, matchmakeextension.MethodFindMatchmakeSessionBySingleGatheringID, callID, stream.Bytes()), nil
}




func (p *Protocol) findMatchmakeSessionByOwner(err error, packet nex.PacketInterface, callID uint32, id types.UInt32, resultRange types.ResultRange) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()

	sessions := types.NewList[matchmakingtypes.MatchmakeSession]()
	for _, session := range p.Store.FindByOwnerLocked(uint64(id)) {
		sessions = append(sessions, publicCopy(session.MatchmakeSession))
	}

	p.Store.RUnlock()

	stream := newStream(endpoint)
	applyRange(&sessions, resultRange).WriteTo(stream)

	return respond(endpoint, matchmakeextension.ProtocolID, matchmakeextension.MethodFindMatchmakeSessionByOwner, callID, stream.Bytes()), nil
}




func (p *Protocol) getPlayingSession(err error, packet nex.PacketInterface, callID uint32, lstPID types.List[types.PID]) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}


	if len(lstPID) > 300 {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "lstPID exceeds MAX_PRINCIPALID_SIZE_TO_FIND_MATCHMAKE_SESSION")
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()

	playingSessions := types.NewList[matchmakingtypes.PlayingSession]()
	for _, pid := range lstPID {
		for _, session := range p.Store.FindByParticipantLocked(uint64(pid)) {
			playingSession := matchmakingtypes.NewPlayingSession()
			playingSession.PrincipalID = pid
			playingSession.Gathering = holderFor(publicCopy(session.MatchmakeSession))

			playingSessions = append(playingSessions, playingSession)
		}
	}

	p.Store.RUnlock()

	stream := newStream(endpoint)
	playingSessions.WriteTo(stream)

	return respond(endpoint, matchmakeextension.ProtocolID, matchmakeextension.MethodGetPlayingSession, callID, stream.Bytes()), nil
}





func (p *Protocol) getSimplePlayingSession(err error, packet nex.PacketInterface, callID uint32, listPID types.List[types.PID], includeLoginUser types.Bool) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	pids := make([]types.PID, 0, len(listPID)+1)
	pids = append(pids, listPID...)

	if bool(includeLoginUser) {
		pids = append(pids, connection.PID())
	}

	p.Store.RLock()

	simpleSessions := types.NewList[matchmakingtypes.SimplePlayingSession]()
	for _, pid := range pids {
		for _, session := range p.Store.FindByParticipantLocked(uint64(pid)) {
			simple := matchmakingtypes.NewSimplePlayingSession()
			simple.PrincipalID = pid
			simple.GatheringID = session.MatchmakeSession.ID
			simple.GameMode = session.MatchmakeSession.GameMode

			if len(session.MatchmakeSession.Attributes) > 0 {
				simple.Attribute0 = session.MatchmakeSession.Attributes[0]
			}

			simpleSessions = append(simpleSessions, simple)
		}
	}

	p.Store.RUnlock()

	stream := newStream(endpoint)
	simpleSessions.WriteTo(stream)

	return respond(endpoint, matchmakeextension.ProtocolID, matchmakeextension.MethodGetSimplePlayingSession, callID, stream.Bytes()), nil
}







func (p *Protocol) ownedSession(connection *nex.PRUDPConnection, gid uint32) (*Session, *nex.Error) {
	session, nexError := p.Store.GetLocked(gid)
	if nexError != nil {
		return nil, nexError
	}

	if !session.MatchmakeSession.OwnerPID.Equals(connection.PID()) {
		return nil, nex.NewError(nex.ResultCodes.RendezVous.PermissionDenied, "Only the gathering owner may modify it")
	}

	return session, nil
}


func (p *Protocol) mutateOwned(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32, methodID uint32, mutate func(session *Session) *nex.Error) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()

	session, nexError := p.ownedSession(connection, uint32(gid))
	if nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	if mutate != nil {
		if nexError = mutate(session); nexError != nil {
			p.Store.Unlock()
			return nil, nexError
		}
	}

	p.Store.Unlock()

	return respondEmpty(endpoint, matchmakeextension.ProtocolID, methodID, callID), nil
}



func (p *Protocol) closeParticipation(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32) (*nex.RMCMessage, *nex.Error) {
	return p.mutateOwned(err, packet, callID, gid, matchmakeextension.MethodCloseParticipation, func(session *Session) *nex.Error {
		session.MatchmakeSession.OpenParticipation = types.NewBool(false)
		return nil
	})
}


func (p *Protocol) openParticipation(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32) (*nex.RMCMessage, *nex.Error) {
	return p.mutateOwned(err, packet, callID, gid, matchmakeextension.MethodOpenParticipation, func(session *Session) *nex.Error {
		session.MatchmakeSession.OpenParticipation = types.NewBool(true)
		return nil
	})
}



func (p *Protocol) updateApplicationBuffer(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32, applicationBuffer types.Buffer) (*nex.RMCMessage, *nex.Error) {
	if len(applicationBuffer) > 512 {
		return nil, invalidArgument(nil)
	}

	return p.mutateOwned(err, packet, callID, gid, matchmakeextension.MethodUpdateApplicationBuffer, func(session *Session) *nex.Error {
		session.MatchmakeSession.ApplicationBuffer = applicationBuffer.Copy().(types.Buffer)
		return nil
	})
}



func (p *Protocol) updateProgressScore(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32, progressScore types.UInt8) (*nex.RMCMessage, *nex.Error) {
	if progressScore > 100 {
		return nil, invalidArgument(nil)
	}

	return p.mutateOwned(err, packet, callID, gid, matchmakeextension.MethodUpdateProgressScore, func(session *Session) *nex.Error {
		session.MatchmakeSession.ProgressScore = progressScore
		return nil
	})
}


func (p *Protocol) updateMatchmakeSessionAttribute(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32, attribs types.List[types.UInt32]) (*nex.RMCMessage, *nex.Error) {
	return p.mutateOwned(err, packet, callID, gid, matchmakeextension.MethodUpdateMatchmakeSessionAttribute, func(session *Session) *nex.Error {
		if len(attribs) != len(session.MatchmakeSession.Attributes) {
			return invalidArgument(nil)
		}

		session.MatchmakeSession.Attributes = attribs.Copy().(types.List[types.UInt32])
		return nil
	})
}


func (p *Protocol) modifyCurrentGameAttribute(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32, attribIndex types.UInt32, newValue types.UInt32) (*nex.RMCMessage, *nex.Error) {
	return p.mutateOwned(err, packet, callID, gid, matchmakeextension.MethodModifyCurrentGameAttribute, func(session *Session) *nex.Error {
		if int(attribIndex) >= len(session.MatchmakeSession.Attributes) {
			return invalidArgument(nil)
		}

		session.MatchmakeSession.Attributes[attribIndex] = newValue
		return nil
	})
}






func (p *Protocol) updateMatchmakeSession(err error, packet nex.PacketInterface, callID uint32, anyGathering matchmakingtypes.GatheringHolder) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	updated, nexError := sessionFromHolder(anyGathering)
	if nexError != nil {
		return nil, nexError
	}

	if !commonglobals.CheckValidMatchmakeSession(updated) {
		return nil, invalidArgument(nil)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()

	session, nexError := p.ownedSession(connection, uint32(updated.ID))
	if nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	current := session.MatchmakeSession

	updated.ID = current.ID
	updated.OwnerPID = current.OwnerPID
	updated.HostPID = current.HostPID
	updated.SessionKey = current.SessionKey
	updated.StartedTime = current.StartedTime
	updated.ParticipationCount = types.NewUInt32(session.ParticipationCount())

	session.MatchmakeSession = updated

	p.Store.Unlock()

	return respondEmpty(endpoint, matchmakeextension.ProtocolID, matchmakeextension.MethodUpdateMatchmakeSession, callID), nil
}







func (p *Protocol) updateNotificationData(err error, packet nex.PacketInterface, callID uint32, uiType notificationsconstants.NotificationCategory, uiParam1 types.UInt64, uiParam2 types.UInt64, strParam types.String) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	if uiType < 101 || uiType > 108 {
		return nil, invalidArgument(nil)
	}

	if utf8.RuneCountInString(string(strParam)) > maxMessageLength {
		return nil, invalidArgument(nil)
	}

	connection, endpoint := connectionOf(packet)

	event := notificationstypes.NewNotificationEvent()
	event.PIDSource = connection.PID()
	event.Type = uiType.Build()
	event.Param1 = uiParam1
	event.Param2 = uiParam2
	event.StrParam = strParam

	p.Store.SetNotificationData(uint64(connection.PID()), event)



	if p.Store.GetUserFriendPIDs != nil {
		var targets []uint64
		for _, friend := range p.Store.GetUserFriendPIDs(uint32(connection.PID())) {
			if endpoint.FindConnectionByPID(uint64(friend)) != nil {
				targets = append(targets, uint64(friend))
			}
		}

		if len(targets) > 0 {
			commonglobals.SendNotificationEvent(endpoint, event, targets)
		}
	}

	return respondEmpty(endpoint, matchmakeextension.ProtocolID, matchmakeextension.MethodUpdateNotificationData, callID), nil
}





func (p *Protocol) getFriendNotificationData(err error, packet nex.PacketInterface, callID uint32, uiType notificationsconstants.NotificationCategorySigned) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	if uiType < 101 || uiType > 108 {
		return nil, invalidArgument(nil)
	}

	events := p.friendNotificationData(packet, []uint32{uint32(uiType.ToUnsigned())})

	_, endpoint := connectionOf(packet)

	stream := newStream(endpoint)
	events.WriteTo(stream)

	return respond(endpoint, matchmakeextension.ProtocolID, matchmakeextension.MethodGetFriendNotificationData, callID, stream.Bytes()), nil
}




func (p *Protocol) getLstFriendNotificationData(err error, packet nex.PacketInterface, callID uint32, lstTypes types.List[notificationsconstants.NotificationCategory]) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	categories := make([]uint32, 0, len(lstTypes))
	for _, category := range lstTypes {
		if category < 101 || category > 108 {
			return nil, invalidArgument(nil)
		}

		categories = append(categories, uint32(category))
	}

	events := p.friendNotificationData(packet, categories)

	_, endpoint := connectionOf(packet)

	stream := newStream(endpoint)
	events.WriteTo(stream)

	return respond(endpoint, matchmakeextension.ProtocolID, matchmakeextension.MethodGetlstFriendNotificationData, callID, stream.Bytes()), nil
}

func (p *Protocol) friendNotificationData(packet nex.PacketInterface, categories []uint32) types.List[notificationstypes.NotificationEvent] {
	connection, _ := connectionOf(packet)

	var friends []uint64
	if p.Store.GetUserFriendPIDs != nil {
		for _, friend := range p.Store.GetUserFriendPIDs(uint32(connection.PID())) {
			friends = append(friends, uint64(friend))
		}
	}

	events := types.NewList[notificationstypes.NotificationEvent]()
	events = append(events, p.Store.NotificationDataForPIDs(friends, categories)...)

	return events
}
