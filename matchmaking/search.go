package matchmaking

import (
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	matchmakingconstants "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/constants"
	matchmakingtypes "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
)












type rangeConstraint struct {
	min uint64
	max uint64
}




func parseConstraint(value string) (constraint rangeConstraint, present bool, ok bool) {
	if value == "" {
		return rangeConstraint{}, false, true
	}

	before, after, found := strings.Cut(value, ",")

	minimum, err := strconv.ParseUint(before, 10, 32)
	if err != nil {
		return rangeConstraint{}, false, false
	}

	if !found {
		return rangeConstraint{min: minimum, max: minimum}, true, true
	}

	maximum, err := strconv.ParseUint(after, 10, 32)
	if err != nil {
		return rangeConstraint{}, false, false
	}

	return rangeConstraint{min: minimum, max: maximum}, true, true
}

func (r rangeConstraint) contains(value uint64) bool {
	return value >= r.min && value <= r.max
}


type SearchOptions struct {


	AutoMatchmake bool



	ExcludeOwnerPID uint64


	AllowPublic bool
}





func (s *Store) Search(
	connection *nex.PRUDPConnection,
	searchCriterias []matchmakingtypes.MatchmakeSessionSearchCriteria,
	resultRange types.ResultRange,
	sourceSession *matchmakingtypes.MatchmakeSession,
	options SearchOptions,
) []*Session {
	results := make([]*Session, 0)

	if !options.AllowPublic {
		return results
	}

	var friendList []uint32
	if s.GetUserFriendPIDs != nil {
		friendList = s.GetUserFriendPIDs(uint32(connection.PID()))
	}

	offset := uint32(resultRange.Offset)
	if offset == math.MaxUint32 {
		offset = 0
	}

	length := uint32(resultRange.Length)
	if length == 0 {
		length = math.MaxUint32
	}

	for _, criteria := range searchCriterias {
		if options.AutoMatchmake {
			criteria.VacantOnly = true
			criteria.ExcludeLocked = true
			criteria.ExcludeNonHostPID = true
		}

		matched, ok := s.matchCriteria(connection, criteria, friendList, options)
		if !ok {


			continue
		}

		sortSessions(matched, criteria, sourceSession)


		if offset >= uint32(len(matched)) {
			continue
		}

		matched = matched[offset:]

		remaining := int(length) - len(results)
		if remaining <= 0 {
			break
		}

		if len(matched) > remaining {
			matched = matched[:remaining]
		}

		results = append(results, matched...)
	}

	return results
}


func (s *Store) matchCriteria(
	connection *nex.PRUDPConnection,
	criteria matchmakingtypes.MatchmakeSessionSearchCriteria,
	friendList []uint32,
	options SearchOptions,
) ([]*Session, bool) {


	attributeConstraints := make([]rangeConstraint, len(criteria.Attribs))
	attributePresent := make([]bool, len(criteria.Attribs))

	for i, attribute := range criteria.Attribs {


		if i == 1 {
			continue
		}

		constraint, present, ok := parseConstraint(string(attribute))
		if !ok {
			return nil, false
		}

		attributeConstraints[i] = constraint
		attributePresent[i] = present
	}

	maxParticipants, maxParticipantsPresent, ok := parseConstraint(string(criteria.MaxParticipants))
	if !ok {
		return nil, false
	}

	minParticipants, minParticipantsPresent, ok := parseConstraint(string(criteria.MinParticipants))
	if !ok {
		return nil, false
	}

	gameMode, gameModePresent, ok := parseConstraint(string(criteria.GameMode))
	if !ok {
		return nil, false
	}

	systemType, systemTypePresent, ok := parseConstraint(string(criteria.MatchmakeSystemType))
	if !ok {
		return nil, false
	}

	vacantParticipants := uint32(criteria.VacantParticipants)
	if vacantParticipants == 0 {
		vacantParticipants = 1
	}

	matched := make([]*Session, 0)

	for _, session := range s.AllLocked() {
		candidate := session.MatchmakeSession

		if options.ExcludeOwnerPID != 0 && uint64(candidate.OwnerPID) == options.ExcludeOwnerPID {
			continue
		}



		if len(candidate.Attributes) != len(criteria.Attribs) {
			continue
		}

		attributesMatch := true
		for i := range criteria.Attribs {
			if !attributePresent[i] {
				continue
			}

			if !attributeConstraints[i].contains(uint64(candidate.Attributes[i])) {
				attributesMatch = false
				break
			}
		}

		if !attributesMatch {
			continue
		}

		if maxParticipantsPresent && !maxParticipants.contains(uint64(candidate.MaximumParticipants)) {
			continue
		}

		if minParticipantsPresent && !minParticipants.contains(uint64(candidate.MinimumParticipants)) {
			continue
		}

		if gameModePresent && !gameMode.contains(uint64(candidate.GameMode)) {
			continue
		}

		if systemTypePresent && !systemType.contains(uint64(candidate.MatchmakeSystemType)) {
			continue
		}


		if candidate.ParticipationPolicy == 98 && !slices.Contains(friendList, uint32(candidate.OwnerPID)) {
			continue
		}

		if bool(criteria.ExcludeLocked) && !bool(candidate.OpenParticipation) {
			continue
		}

		if bool(criteria.ExcludeNonHostPID) && uint64(candidate.HostPID) == 0 {
			continue
		}

		if bool(criteria.VacantOnly) {
			maximum := uint32(candidate.MaximumParticipants)
			if maximum != 0 && session.ParticipationCount()+vacantParticipants > maximum {
				continue
			}
		}

		matched = append(matched, session)
	}

	return matched, true
}


func sortSessions(sessions []*Session, criteria matchmakingtypes.MatchmakeSessionSearchCriteria, sourceSession *matchmakingtypes.MatchmakeSession) {
	distanceTo := func(target uint64, value uint64) uint64 {
		if value > target {
			return value - target
		}

		return target - value
	}


	referenceAttribute := uint64(0)
	if len(criteria.Attribs) > 1 {
		if parsed, err := strconv.ParseUint(string(criteria.Attribs[1]), 10, 32); err == nil {
			referenceAttribute = parsed
		}
	}

	attributeOf := func(session *Session) uint64 {
		if len(session.MatchmakeSession.Attributes) > 1 {
			return uint64(session.MatchmakeSession.Attributes[1])
		}

		return 0
	}

	switch criteria.SelectionMethod {
	case matchmakingconstants.MatchmakeSelectionMethodNearestNeighbor,
		matchmakingconstants.MatchmakeSelectionMethodBroadenRange:
		sort.SliceStable(sessions, func(a, b int) bool {
			return distanceTo(referenceAttribute, attributeOf(sessions[a])) <
				distanceTo(referenceAttribute, attributeOf(sessions[b]))
		})

	case matchmakingconstants.MatchmakeSelectionMethodProgressScore:
		if sourceSession == nil {
			return
		}

		target := uint64(sourceSession.ProgressScore)
		sort.SliceStable(sessions, func(a, b int) bool {
			return distanceTo(target, uint64(sessions[a].MatchmakeSession.ProgressScore)) <
				distanceTo(target, uint64(sessions[b].MatchmakeSession.ProgressScore))
		})

	case matchmakingconstants.MatchmakeSelectionMethodBroadenRangeWithProgressScore:
		if sourceSession == nil {
			return
		}

		target := uint64(sourceSession.ProgressScore)
		sort.SliceStable(sessions, func(a, b int) bool {
			scoreA := distanceTo(referenceAttribute, attributeOf(sessions[a])) + distanceTo(target, uint64(sessions[a].MatchmakeSession.ProgressScore))
			scoreB := distanceTo(referenceAttribute, attributeOf(sessions[b])) + distanceTo(target, uint64(sessions[b].MatchmakeSession.ProgressScore))
			return scoreA < scoreB
		})

	default:




		sort.SliceStable(sessions, func(a, b int) bool {
			return sessions[a].ParticipationCount() > sessions[b].ParticipationCount()
		})
	}
}














func (s *Store) FindSession(connection *nex.PRUDPConnection, search matchmakingtypes.MatchmakeSession, allowPublic bool) *Session {
	if !allowPublic {
		return nil
	}

	var friendList []uint32
	if s.GetUserFriendPIDs != nil {
		friendList = s.GetUserFriendPIDs(uint32(connection.PID()))
	}



	matchedAttributes := func(candidate matchmakingtypes.MatchmakeSession) bool {
		if len(candidate.Attributes) != len(search.Attributes) {
			return false
		}

		for i := range search.Attributes {

			if i == 1 {
				continue
			}

			if candidate.Attributes[i] != search.Attributes[i] {
				return false
			}
		}

		return true
	}

	candidates := make([]*Session, 0)

	for _, session := range s.AllLocked() {
		candidate := session.MatchmakeSession

		if uint64(candidate.HostPID) == 0 {
			continue
		}

		if !bool(candidate.OpenParticipation) {
			continue
		}

		if bool(candidate.UserPasswordEnabled) || bool(candidate.SystemPasswordEnabled) {
			continue
		}

		maximum := uint32(candidate.MaximumParticipants)
		if maximum != 0 && session.ParticipationCount() >= maximum {
			continue
		}

		if candidate.MaximumParticipants != search.MaximumParticipants {
			continue
		}

		if candidate.MinimumParticipants != search.MinimumParticipants {
			continue
		}

		if candidate.GameMode != search.GameMode {
			continue
		}

		if candidate.MatchmakeSystemType != search.MatchmakeSystemType {
			continue
		}

		if !matchedAttributes(candidate) {
			continue
		}

		if candidate.ParticipationPolicy == 98 && !slices.Contains(friendList, uint32(candidate.OwnerPID)) {
			continue
		}

		candidates = append(candidates, session)
	}

	if len(candidates) == 0 {
		return nil
	}

	reference := uint64(0)
	if len(search.Attributes) > 1 {
		reference = uint64(search.Attributes[1])
	}

	distance := func(session *Session) uint64 {
		value := uint64(0)
		if len(session.MatchmakeSession.Attributes) > 1 {
			value = uint64(session.MatchmakeSession.Attributes[1])
		}

		if value > reference {
			return value - reference
		}

		return reference - value
	}

	sort.SliceStable(candidates, func(a, b int) bool {
		return distance(candidates[a]) < distance(candidates[b])
	})

	return candidates[0]
}


func (s *Store) CanJoin(pid types.PID, session *Session) *nex.Error {
	matchmakeSession := session.MatchmakeSession

	if !bool(matchmakeSession.OpenParticipation) {
		return nex.NewError(nex.ResultCodes.RendezVous.PermissionDenied, "Gathering is not open to new participants")
	}

	if matchmakeSession.ParticipationPolicy == 98 {
		if s.GetUserFriendPIDs == nil {
			return nex.NewError(nex.ResultCodes.Core.NotImplemented, "Friend lookups are unavailable")
		}

		friendList := s.GetUserFriendPIDs(uint32(pid))
		if !slices.Contains(friendList, uint32(matchmakeSession.OwnerPID)) {
			return nex.NewError(nex.ResultCodes.RendezVous.NotFriend, "Not a friend of the gathering owner")
		}
	}

	return nil
}
