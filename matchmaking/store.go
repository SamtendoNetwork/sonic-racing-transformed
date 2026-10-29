









package matchmaking

import (
	"crypto/rand"
	"sync"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	commonglobals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	matchmakingconstants "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/constants"
	matchmakingtypes "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
	notificationsconstants "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/constants"
	notificationstypes "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/types"
)


type Session struct {


	MatchmakeSession matchmakingtypes.MatchmakeSession





	Participants []uint64


	StartedTime types.DateTime


	SessionURL types.String
}


func (s *Session) ParticipationCount() uint32 {
	return uint32(len(s.Participants))
}


func (s *Session) Contains(pid uint64) bool {
	for _, participant := range s.Participants {
		if participant == pid {
			return true
		}
	}

	return false
}


type Store struct {
	mutex    sync.RWMutex
	sessions map[uint32]*Session



	nextGatheringID uint32




	notificationData map[uint64]map[uint32]notificationstypes.NotificationEvent



	Endpoint *nex.PRUDPEndPoint




	GetUserFriendPIDs func(pid uint32) []uint32
}


func NewStore(endpoint *nex.PRUDPEndPoint) *Store {
	return &Store{
		sessions:         make(map[uint32]*Session),
		notificationData: make(map[uint64]map[uint32]notificationstypes.NotificationEvent),
		nextGatheringID:  1,
		Endpoint:         endpoint,
		GetUserFriendPIDs: func(pid uint32) []uint32 {
			return []uint32{}
		},
	}
}



func (s *Store) Lock()    { s.mutex.Lock() }
func (s *Store) Unlock()  { s.mutex.Unlock() }
func (s *Store) RLock()   { s.mutex.RLock() }
func (s *Store) RUnlock() { s.mutex.RUnlock() }



func (s *Store) GetLocked(gatheringID uint32) (*Session, *nex.Error) {
	session, ok := s.sessions[gatheringID]
	if !ok {
		return nil, nex.NewError(nex.ResultCodes.RendezVous.SessionVoid, "Gathering does not exist")
	}

	return session, nil
}


func (s *Store) Get(gatheringID uint32) (*Session, *nex.Error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.GetLocked(gatheringID)
}


func (s *Store) AllLocked() []*Session {
	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}







func (s *Store) CreateLocked(connection *nex.PRUDPConnection, session matchmakingtypes.MatchmakeSession) *Session {
	gatheringID := s.nextGatheringID
	s.nextGatheringID++

	session.ID = types.NewUInt32(gatheringID)
	session.OwnerPID = connection.PID()
	session.HostPID = connection.PID()
	session.StartedTime = types.NewDateTime(0).Now()




	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {


		commonglobals.Logger.Errorf("Failed to generate a session key for gathering %d: %v", gatheringID, err)
	}

	session.SessionKey = types.NewBuffer(sessionKey)
	session.UserPasswordEnabled = types.NewBool(len(session.UserPassword) > 0)
	session.SystemPasswordEnabled = types.NewBool(false)

	stored := &Session{
		MatchmakeSession: session,
		Participants:     make([]uint64, 0, session.MaximumParticipants),
		StartedTime:      types.NewDateTime(0).Now(),
	}

	s.sessions[gatheringID] = stored

	return stored
}






func (s *Store) JoinLocked(session *Session, connection *nex.PRUDPConnection, vacantParticipants uint16, joinMessage string) (uint32, *nex.Error) {
	if vacantParticipants == 0 {
		vacantParticipants = 1
	}

	pid := uint64(connection.PID())

	if session.Contains(pid) {
		return 0, nex.NewError(nex.ResultCodes.RendezVous.AlreadyParticipatedGathering, "Already a participant of this gathering")
	}

	maximumParticipants := uint32(session.MatchmakeSession.MaximumParticipants)
	if maximumParticipants != 0 && session.ParticipationCount()+uint32(vacantParticipants) > maximumParticipants {
		return 0, nex.NewError(nex.ResultCodes.RendezVous.SessionFull, "Gathering is full")
	}

	previousParticipants := make([]uint64, len(session.Participants))
	copy(previousParticipants, session.Participants)

	for i := uint16(0); i < vacantParticipants; i++ {
		session.Participants = append(session.Participants, pid)
	}

	total := session.ParticipationCount()
	session.MatchmakeSession.ParticipationCount = types.NewUInt32(total)

	gatheringID := uint32(session.MatchmakeSession.ID)
	flags := session.MatchmakeSession.Flags

	var targets []uint64

	if flags.HasFlag(matchmakingconstants.GatheringFlagNotifyParticipationEventsToAllParticipants) ||
		flags.HasFlag(matchmakingconstants.GatheringFlagNotifyParticipationEventsToAllParticipantsReproducibly) {
		targets = commonglobals.RemoveDuplicates(session.Participants)
	} else {


		if pid == uint64(session.MatchmakeSession.OwnerPID) {
			return total, nil
		}

		targets = []uint64{uint64(session.MatchmakeSession.OwnerPID)}
	}

	s.notifyParticipation(
		connection.PID(),
		notificationsconstants.ParticipationEventsParticipate,
		gatheringID,
		pid,
		joinMessage,
		uint64(total),
		targets,
	)



	if flags.HasFlag(matchmakingconstants.GatheringFlagNotifyParticipationEventsToAllParticipantsReproducibly) {
		for _, participant := range commonglobals.RemoveDuplicates(previousParticipants) {
			s.notifyParticipation(
				connection.PID(),
				notificationsconstants.ParticipationEventsParticipate,
				gatheringID,
				participant,
				joinMessage,
				uint64(total),
				[]uint64{pid},
			)
		}
	}

	return total, nil
}









func (s *Store) LeaveLocked(session *Session, pid uint64, subType notificationsconstants.ParticipationEvents, message string) {
	if !session.Contains(pid) {
		return
	}

	remaining := make([]uint64, 0, len(session.Participants))
	for _, participant := range session.Participants {
		if participant != pid {
			remaining = append(remaining, participant)
		}
	}

	session.Participants = remaining
	session.MatchmakeSession.ParticipationCount = types.NewUInt32(session.ParticipationCount())

	gatheringID := uint32(session.MatchmakeSession.ID)

	s.notifyParticipation(
		types.NewPID(pid),
		subType,
		gatheringID,
		pid,
		message,
		uint64(session.ParticipationCount()),
		commonglobals.RemoveDuplicates(session.Participants),
	)

	if uint64(session.MatchmakeSession.OwnerPID) == pid || len(session.Participants) == 0 {
		delete(s.sessions, gatheringID)
		return
	}

	if uint64(session.MatchmakeSession.HostPID) == pid {
		session.MatchmakeSession.HostPID = session.MatchmakeSession.OwnerPID
	}
}




func (s *Store) RemoveConnection(pid uint64, subType notificationsconstants.ParticipationEvents) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.RemoveConnectionLocked(pid, subType)
}


func (s *Store) RemoveConnectionLocked(pid uint64, subType notificationsconstants.ParticipationEvents) {
	for _, session := range s.AllLocked() {
		if session.Contains(pid) {
			s.LeaveLocked(session, pid, subType, "")
		}
	}
}






func (s *Store) notifyParticipation(source types.PID, subType notificationsconstants.ParticipationEvents, gatheringID uint32, subjectPID uint64, message string, participantCount uint64, targets []uint64) {
	if s.Endpoint == nil || len(targets) == 0 {
		return
	}

	event := notificationstypes.NewNotificationEvent()
	event.PIDSource = source
	event.Type = notificationsconstants.NotificationCategoryParticipationEvent.Build(subType)
	event.Param1 = types.UInt64(gatheringID)
	event.Param2 = types.UInt64(subjectPID)
	event.StrParam = types.NewString(message)
	event.Param3 = types.UInt64(participantCount)

	commonglobals.SendNotificationEvent(s.Endpoint, event, targets)
}


func (s *Store) SetNotificationData(pid uint64, event notificationstypes.NotificationEvent) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.notificationData[pid]; !ok {
		s.notificationData[pid] = make(map[uint32]notificationstypes.NotificationEvent)
	}

	s.notificationData[pid][uint32(event.Type)] = event
}



func (s *Store) NotificationDataForPIDs(pids []uint64, categories []uint32) []notificationstypes.NotificationEvent {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	events := make([]notificationstypes.NotificationEvent, 0)

	for _, pid := range pids {
		stored, ok := s.notificationData[pid]
		if !ok {
			continue
		}

		for _, category := range categories {
			for eventType, event := range stored {


				if eventType/1000 == category || eventType == category {
					events = append(events, event)
				}
			}
		}
	}

	return events
}



func (s *Store) FindByParticipantLocked(pid uint64) []*Session {
	found := make([]*Session, 0)

	for _, session := range s.AllLocked() {
		if session.Contains(pid) {
			found = append(found, session)
		}
	}

	return found
}



func (s *Store) FindByOwnerLocked(pid uint64) []*Session {
	found := make([]*Session, 0)

	for _, session := range s.AllLocked() {
		if uint64(session.MatchmakeSession.OwnerPID) == pid {
			found = append(found, session)
		}
	}

	return found
}
