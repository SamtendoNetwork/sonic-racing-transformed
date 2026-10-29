package matchmaking

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	commonglobals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	matchmakingprotocol "github.com/PretendoNetwork/nex-protocols-go/v2/match-making"
	matchmakingconstants "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/constants"
	matchmakingtypes "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
	notificationsconstants "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/constants"
	notificationstypes "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/types"
)







func (p *Protocol) registerMatchMaking(protocol *matchmakingprotocol.Protocol) {
	protocol.SetHandlerRegisterGathering(p.registerGathering)
	protocol.SetHandlerUnregisterGathering(p.unregisterGathering)
	protocol.SetHandlerUnregisterGatherings(p.unregisterGatherings)
	protocol.SetHandlerUpdateGathering(p.updateGathering)
	protocol.SetHandlerParticipate(p.participate)
	protocol.SetHandlerCancelParticipation(p.cancelParticipation)
	protocol.SetHandlerGetParticipants(p.getParticipants)
	protocol.SetHandlerGetDetailedParticipants(p.getDetailedParticipants)
	protocol.SetHandlerGetParticipantsURLs(p.getParticipantsURLs)
	protocol.SetHandlerFindByID(p.findByID)
	protocol.SetHandlerFindBySingleID(p.findBySingleID)
	protocol.SetHandlerFindByOwner(p.findByOwner)
	protocol.SetHandlerFindByParticipants(p.findByParticipants)
	protocol.SetHandlerLaunchSession(p.launchSession)
	protocol.SetHandlerUpdateSessionURL(p.updateSessionURL)
	protocol.SetHandlerGetSessionURL(p.getSessionURL)
	protocol.SetHandlerGetSessionURLs(p.getSessionURLs)
	protocol.SetHandlerGetState(p.getState)
	protocol.SetHandlerSetState(p.setState)
	protocol.SetHandlerUpdateSessionHost(p.updateSessionHost)
	protocol.SetHandlerUpdateSessionHostV1(p.updateSessionHostV1)
	protocol.SetHandlerMigrateGatheringOwnershipV1(p.migrateGatheringOwnershipV1)
}




func (p *Protocol) registerGathering(err error, packet nex.PacketInterface, callID uint32, anyGathering matchmakingtypes.GatheringHolder) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	session, nexError := sessionFromHolder(anyGathering)
	if nexError != nil {
		return nil, nexError
	}

	if !commonglobals.CheckValidMatchmakeSession(session) {
		return nil, invalidArgument(nil)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()

	p.Store.RemoveConnectionLocked(uint64(connection.PID()), notificationsconstants.ParticipationEventsDisconnect)

	stored := p.Store.CreateLocked(connection, session)

	if _, nexError = p.Store.JoinLocked(stored, connection, 1, ""); nexError != nil {
		delete(p.Store.sessions, uint32(stored.MatchmakeSession.ID))
		p.Store.Unlock()
		return nil, nexError
	}

	gatheringID := stored.MatchmakeSession.ID

	p.Store.Unlock()

	p.Logger.Infof("Gathering %d registered by PID %d", uint32(gatheringID), uint64(connection.PID()))

	stream := newStream(endpoint)
	gatheringID.WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodRegisterGathering, callID, stream.Bytes()), nil
}





func (p *Protocol) unregisterGathering(err error, packet nex.PacketInterface, callID uint32, idGathering types.UInt32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	nexError := p.destroyGathering(connection, endpoint, uint32(idGathering))
	if nexError != nil {
		return nil, nexError
	}

	stream := newStream(endpoint)
	types.NewBool(true).WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodUnregisterGathering, callID, stream.Bytes()), nil
}




func (p *Protocol) unregisterGatherings(err error, packet nex.PacketInterface, callID uint32, lstGatherings types.List[types.UInt32]) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	for _, gatheringID := range lstGatherings {
		if nexError := p.destroyGathering(connection, endpoint, uint32(gatheringID)); nexError != nil {
			return nil, nexError
		}
	}

	stream := newStream(endpoint)
	types.NewBool(true).WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodUnregisterGatherings, callID, stream.Bytes()), nil
}

func (p *Protocol) destroyGathering(connection *nex.PRUDPConnection, endpoint *nex.PRUDPEndPoint, gatheringID uint32) *nex.Error {
	p.Store.Lock()

	session, nexError := p.Store.GetLocked(gatheringID)
	if nexError != nil {
		p.Store.Unlock()
		return nexError
	}

	if !session.MatchmakeSession.OwnerPID.Equals(connection.PID()) {
		p.Store.Unlock()
		return nex.NewError(nex.ResultCodes.RendezVous.PermissionDenied, "Only the gathering owner may unregister it")
	}

	participants := commonglobals.RemoveDuplicates(session.Participants)
	delete(p.Store.sessions, gatheringID)

	p.Store.Unlock()

	event := notificationstypes.NewNotificationEvent()
	event.PIDSource = connection.PID()
	event.Type = notificationsconstants.NotificationCategoryGatheringUnregistered.Build()
	event.Param1 = types.UInt64(gatheringID)

	commonglobals.SendNotificationEvent(endpoint, event, participants)

	p.Logger.Infof("Gathering %d unregistered by PID %d", gatheringID, uint64(connection.PID()))

	return nil
}




func (p *Protocol) updateGathering(err error, packet nex.PacketInterface, callID uint32, anyGathering matchmakingtypes.GatheringHolder) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	updated, nexError := sessionFromHolder(anyGathering)
	if nexError != nil {
		return nil, nexError
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

	stream := newStream(endpoint)
	types.NewBool(true).WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodUpdateGathering, callID, stream.Bytes()), nil
}




func (p *Protocol) participate(err error, packet nex.PacketInterface, callID uint32, idGathering types.UInt32, strMessage types.String) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	if len(strMessage) > maxMessageLength {
		return nil, invalidArgument(nil)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()

	session, nexError := p.Store.GetLocked(uint32(idGathering))
	if nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	if nexError = p.Store.CanJoin(connection.PID(), session); nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	if _, nexError = p.Store.JoinLocked(session, connection, 1, string(strMessage)); nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	p.Store.Unlock()

	stream := newStream(endpoint)
	types.NewBool(true).WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodParticipate, callID, stream.Bytes()), nil
}




func (p *Protocol) cancelParticipation(err error, packet nex.PacketInterface, callID uint32, idGathering types.UInt32, strMessage types.String) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	if len(strMessage) > maxMessageLength {
		return nil, invalidArgument(nil)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()

	session, nexError := p.Store.GetLocked(uint32(idGathering))
	if nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	p.Store.LeaveLocked(session, uint64(connection.PID()), notificationsconstants.ParticipationEventsCancelParticipation, string(strMessage))

	p.Store.Unlock()

	stream := newStream(endpoint)
	types.NewBool(true).WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodCancelParticipation, callID, stream.Bytes()), nil
}




func (p *Protocol) getParticipants(err error, packet nex.PacketInterface, callID uint32, idGathering types.UInt32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()
	session, nexError := p.Store.GetLocked(uint32(idGathering))
	if nexError != nil {
		p.Store.RUnlock()
		return nil, nexError
	}

	participants := participantList(session)
	p.Store.RUnlock()

	stream := newStream(endpoint)
	participants.WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodGetParticipants, callID, stream.Bytes()), nil
}




func (p *Protocol) getDetailedParticipants(err error, packet nex.PacketInterface, callID uint32, idGathering types.UInt32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()
	session, nexError := p.Store.GetLocked(uint32(idGathering))
	if nexError != nil {
		p.Store.RUnlock()
		return nil, nexError
	}

	details := participantDetails(session)
	p.Store.RUnlock()

	stream := newStream(endpoint)
	details.WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodGetDetailedParticipants, callID, stream.Bytes()), nil
}









func (p *Protocol) getParticipantsURLs(err error, packet nex.PacketInterface, callID uint32, idGathering types.UInt32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.RLock()
	session, nexError := p.Store.GetLocked(uint32(idGathering))
	if nexError != nil {
		p.Store.RUnlock()
		return nil, nexError
	}

	if !session.Contains(uint64(connection.PID())) {
		p.Store.RUnlock()
		return nil, nex.NewError(nex.ResultCodes.RendezVous.PermissionDenied, "Not a participant of this gathering")
	}

	urls := stationURLsOf(endpoint, session)
	p.Store.RUnlock()

	stream := newStream(endpoint)
	urls.WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodGetParticipantsURLs, callID, stream.Bytes()), nil
}




func (p *Protocol) findByID(err error, packet nex.PacketInterface, callID uint32, lstID types.List[types.UInt32]) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()

	gatherings := types.NewList[matchmakingtypes.GatheringHolder]()
	for _, gatheringID := range lstID {
		if session, nexError := p.Store.GetLocked(uint32(gatheringID)); nexError == nil {
			gatherings = append(gatherings, holderFor(publicCopy(session.MatchmakeSession)))
		}
	}

	p.Store.RUnlock()

	stream := newStream(endpoint)
	gatherings.WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodFindByID, callID, stream.Bytes()), nil
}




func (p *Protocol) findBySingleID(err error, packet nex.PacketInterface, callID uint32, id types.UInt32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()
	session, nexError := p.Store.GetLocked(uint32(id))

	found := types.NewBool(nexError == nil)
	holder := matchmakingtypes.NewGatheringHolder()
	if nexError == nil {
		holder = holderFor(publicCopy(session.MatchmakeSession))
	} else {
		holder.Object = matchmakingtypes.NewMatchmakeSession()
	}
	p.Store.RUnlock()

	stream := newStream(endpoint)
	found.WriteTo(stream)
	holder.WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodFindBySingleID, callID, stream.Bytes()), nil
}




func (p *Protocol) findByOwner(err error, packet nex.PacketInterface, callID uint32, id types.PID, resultRange types.ResultRange) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()

	gatherings := types.NewList[matchmakingtypes.GatheringHolder]()
	for _, session := range p.Store.FindByOwnerLocked(uint64(id)) {
		gatherings = append(gatherings, holderFor(publicCopy(session.MatchmakeSession)))
	}

	p.Store.RUnlock()

	stream := newStream(endpoint)
	applyRange(&gatherings, resultRange).WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodFindByOwner, callID, stream.Bytes()), nil
}




func (p *Protocol) findByParticipants(err error, packet nex.PacketInterface, callID uint32, pid types.List[types.PID]) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()

	seen := make(map[uint32]bool)
	gatherings := types.NewList[matchmakingtypes.GatheringHolder]()

	for _, participant := range pid {
		for _, session := range p.Store.FindByParticipantLocked(uint64(participant)) {
			gatheringID := uint32(session.MatchmakeSession.ID)
			if seen[gatheringID] {
				continue
			}

			seen[gatheringID] = true
			gatherings = append(gatherings, holderFor(publicCopy(session.MatchmakeSession)))
		}
	}

	p.Store.RUnlock()

	stream := newStream(endpoint)
	gatherings.WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodFindByParticipants, callID, stream.Bytes()), nil
}




func (p *Protocol) launchSession(err error, packet nex.PacketInterface, callID uint32, idGathering types.UInt32, strURL types.String) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()

	session, nexError := p.participantSession(connection, uint32(idGathering))
	if nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	session.SessionURL = strURL
	session.MatchmakeSession.HostPID = connection.PID()

	p.Store.Unlock()

	stream := newStream(endpoint)
	types.NewBool(true).WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodLaunchSession, callID, stream.Bytes()), nil
}







func (p *Protocol) updateSessionURL(err error, packet nex.PacketInterface, callID uint32, idGathering types.UInt32, strURL types.String) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()

	session, nexError := p.participantSession(connection, uint32(idGathering))
	if nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	if strURL != "" {
		session.SessionURL = strURL
	}

	session.MatchmakeSession.HostPID = connection.PID()

	p.Store.Unlock()

	stream := newStream(endpoint)
	types.NewBool(true).WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodUpdateSessionURL, callID, stream.Bytes()), nil
}




func (p *Protocol) getSessionURL(err error, packet nex.PacketInterface, callID uint32, idGathering types.UInt32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()
	session, nexError := p.Store.GetLocked(uint32(idGathering))
	if nexError != nil {
		p.Store.RUnlock()
		return nil, nexError
	}

	sessionURL := session.SessionURL
	p.Store.RUnlock()

	stream := newStream(endpoint)
	types.NewBool(true).WriteTo(stream)
	sessionURL.WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodGetSessionURL, callID, stream.Bytes()), nil
}





func (p *Protocol) getSessionURLs(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.RLock()

	session, nexError := p.Store.GetLocked(uint32(gid))
	if nexError != nil {
		p.Store.RUnlock()
		return nil, nexError
	}

	if !session.Contains(uint64(connection.PID())) {
		p.Store.RUnlock()
		return nil, nex.NewError(nex.ResultCodes.RendezVous.PermissionDenied, "Not a participant of this gathering")
	}

	host := endpoint.FindConnectionByPID(uint64(session.MatchmakeSession.HostPID))

	p.Store.RUnlock()

	urls := types.NewList[types.StationURL]()
	if host != nil {
		urls = host.StationURLs
	} else {
		p.Logger.Warningf("Host of gathering %d is not connected, returning no station URLs", uint32(gid))
	}

	stream := newStream(endpoint)
	urls.WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodGetSessionURLs, callID, stream.Bytes()), nil
}




func (p *Protocol) getState(err error, packet nex.PacketInterface, callID uint32, idGathering types.UInt32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	_, endpoint := connectionOf(packet)

	p.Store.RLock()
	session, nexError := p.Store.GetLocked(uint32(idGathering))
	if nexError != nil {
		p.Store.RUnlock()
		return nil, nexError
	}

	state := types.NewUInt32(uint32(session.MatchmakeSession.State))
	p.Store.RUnlock()

	stream := newStream(endpoint)
	types.NewBool(true).WriteTo(stream)
	state.WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodGetState, callID, stream.Bytes()), nil
}




func (p *Protocol) setState(err error, packet nex.PacketInterface, callID uint32, idGathering types.UInt32, uiNewState matchmakingconstants.GatheringState) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()

	session, nexError := p.ownedSession(connection, uint32(idGathering))
	if nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	session.MatchmakeSession.State = uiNewState

	p.Store.Unlock()

	stream := newStream(endpoint)
	types.NewBool(true).WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodSetState, callID, stream.Bytes()), nil
}







func (p *Protocol) updateSessionHost(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32, isMigrateOwner types.Bool) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()

	session, nexError := p.participantSession(connection, uint32(gid))
	if nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	if !bool(isMigrateOwner) {
		session.MatchmakeSession.HostPID = connection.PID()
		p.Store.Unlock()

		return respondEmpty(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodUpdateSessionHost, callID), nil
	}

	if !session.MatchmakeSession.Flags.HasFlag(matchmakingconstants.GatheringFlagChangeOwnerByOtherHost) {
		p.Store.Unlock()
		return nil, nex.NewError(nex.ResultCodes.RendezVous.InvalidOperation, "Gathering does not permit ownership migration")
	}

	session.MatchmakeSession.HostPID = connection.PID()
	session.MatchmakeSession.OwnerPID = connection.PID()

	participants := commonglobals.RemoveDuplicates(session.Participants)

	p.Store.Unlock()

	event := notificationstypes.NewNotificationEvent()
	event.PIDSource = connection.PID()
	event.Type = notificationsconstants.NotificationCategoryOwnershipChangeEvent.Build()
	event.Param1 = types.UInt64(gid)
	event.Param2 = types.UInt64(connection.PID())

	commonglobals.SendNotificationEvent(endpoint, event, participants)

	p.Logger.Infof("PID %d took ownership of gathering %d", uint64(connection.PID()), uint32(gid))

	return respondEmpty(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodUpdateSessionHost, callID), nil
}




func (p *Protocol) updateSessionHostV1(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, invalidArgument(err)
	}

	connection, endpoint := connectionOf(packet)

	p.Store.Lock()

	session, nexError := p.participantSession(connection, uint32(gid))
	if nexError != nil {
		p.Store.Unlock()
		return nil, nexError
	}

	session.MatchmakeSession.HostPID = connection.PID()

	p.Store.Unlock()

	return respondEmpty(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodUpdateSessionHostV1, callID), nil
}





func (p *Protocol) migrateGatheringOwnershipV1(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32, lstPotentialNewOwnersID types.List[types.PID]) (*nex.RMCMessage, *nex.Error) {
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

	var newOwner types.PID
	for _, candidate := range lstPotentialNewOwnersID {
		if session.Contains(uint64(candidate)) {
			newOwner = candidate
			break
		}
	}

	if newOwner == 0 {
		p.Store.Unlock()
		return nil, nex.NewError(nex.ResultCodes.RendezVous.InvalidOperation, "No candidate owner is a participant")
	}

	session.MatchmakeSession.OwnerPID = newOwner
	participants := commonglobals.RemoveDuplicates(session.Participants)

	p.Store.Unlock()

	event := notificationstypes.NewNotificationEvent()
	event.PIDSource = connection.PID()
	event.Type = notificationsconstants.NotificationCategoryOwnershipChangeEvent.Build()
	event.Param1 = types.UInt64(gid)
	event.Param2 = types.UInt64(newOwner)

	commonglobals.SendNotificationEvent(endpoint, event, participants)

	stream := newStream(endpoint)
	types.NewBool(true).WriteTo(stream)

	return respond(endpoint, matchmakingprotocol.ProtocolID, matchmakingprotocol.MethodMigrateGatheringOwnershipV1, callID, stream.Bytes()), nil
}







func (p *Protocol) participantSession(connection *nex.PRUDPConnection, gatheringID uint32) (*Session, *nex.Error) {
	session, nexError := p.Store.GetLocked(gatheringID)
	if nexError != nil {
		return nil, nexError
	}

	if !session.Contains(uint64(connection.PID())) {
		return nil, nex.NewError(nex.ResultCodes.RendezVous.PermissionDenied, "Not a participant of this gathering")
	}

	return session, nil
}


func participantList(session *Session) types.List[types.PID] {
	participants := types.NewList[types.PID]()
	for _, pid := range commonglobals.RemoveDuplicates(session.Participants) {
		participants = append(participants, types.NewPID(pid))
	}

	return participants
}



func participantDetails(session *Session) types.List[matchmakingtypes.ParticipantDetails] {
	counts := make(map[uint64]uint16)
	for _, pid := range session.Participants {
		counts[pid]++
	}

	details := types.NewList[matchmakingtypes.ParticipantDetails]()
	for _, pid := range commonglobals.RemoveDuplicates(session.Participants) {
		detail := matchmakingtypes.NewParticipantDetails()
		detail.IDParticipant = types.NewPID(pid)


		detail.StrName = types.NewString("")
		detail.StrMessage = types.NewString("")
		detail.UIParticipants = types.NewUInt16(counts[pid])

		details = append(details, detail)
	}

	return details
}


func stationURLsOf(endpoint *nex.PRUDPEndPoint, session *Session) types.List[types.StationURL] {
	urls := types.NewList[types.StationURL]()

	for _, pid := range commonglobals.RemoveDuplicates(session.Participants) {
		participant := endpoint.FindConnectionByPID(pid)
		if participant == nil {
			continue
		}

		urls = append(urls, participant.StationURLs...)
	}

	return urls
}


func applyRange[T types.RVType](list *types.List[T], resultRange types.ResultRange) types.List[T] {
	offset := int(resultRange.Offset)
	length := int(resultRange.Length)

	if offset >= len(*list) {
		return types.NewList[T]()
	}

	sliced := (*list)[offset:]

	if length > 0 && length < len(sliced) {
		sliced = sliced[:length]
	}

	return sliced
}
