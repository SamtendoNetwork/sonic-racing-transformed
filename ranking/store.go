







package ranking

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/PretendoNetwork/nex-go/v2/types"
	rankingconstants "github.com/PretendoNetwork/nex-protocols-go/v2/ranking/constants"
	rankingtypes "github.com/PretendoNetwork/nex-protocols-go/v2/ranking/types"
)


type Entry struct {
	PID      uint64 `json:"pid"`
	UniqueID uint64 `json:"uniqueId"`
	Category uint32 `json:"category"`
	Score    uint32 `json:"score"`




	OrderBy uint8 `json:"orderBy"`



	Groups []byte `json:"groups,omitempty"`



	Param uint64 `json:"param"`


	UpdateTime uint64 `json:"updateTime"`
}


type persistedState struct {

	Scores map[string]map[string]*Entry `json:"scores"`


	CommonData map[string][]byte `json:"commonData"`
}


type Store struct {
	mutex sync.RWMutex




	scores map[uint32]map[uint64]*Entry


	commonData map[uint64][]byte

	path string




	GetUserFriendPIDs func(pid uint32) []uint32
}




func perPlayerCommonDataKey(pid uint64, uniqueID uint64) uint64 {
	if uniqueID == 0 {
		return (uint64(1) << 63) | pid
	}

	return uniqueID
}



func NewStore(dataDir string) (*Store, error) {
	store := &Store{
		scores:     make(map[uint32]map[uint64]*Entry),
		commonData: make(map[uint64][]byte),
		path:       filepath.Join(dataDir, "ranking.json"),
		GetUserFriendPIDs: func(pid uint32) []uint32 {
			return []uint32{}
		},
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) load() error {
	contents, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to read %s: %w", s.path, err)
	}

	var state persistedState
	if err := json.Unmarshal(contents, &state); err != nil {
		return fmt.Errorf("failed to parse %s: %w", s.path, err)
	}

	for categoryKey, entries := range state.Scores {
		var category uint32
		if _, err := fmt.Sscanf(categoryKey, "%d", &category); err != nil {
			continue
		}

		s.scores[category] = make(map[uint64]*Entry, len(entries))

		for pidKey, entry := range entries {
			var pid uint64
			if _, err := fmt.Sscanf(pidKey, "%d", &pid); err != nil {
				continue
			}

			s.scores[category][pid] = entry
		}
	}

	for uniqueIDKey, blob := range state.CommonData {
		var uniqueID uint64
		if _, err := fmt.Sscanf(uniqueIDKey, "%d", &uniqueID); err != nil {
			continue
		}

		s.commonData[uniqueID] = blob
	}

	return nil
}





func (s *Store) save() error {
	state := persistedState{
		Scores:     make(map[string]map[string]*Entry, len(s.scores)),
		CommonData: make(map[string][]byte, len(s.commonData)),
	}

	for category, entries := range s.scores {
		key := fmt.Sprintf("%d", category)
		state.Scores[key] = make(map[string]*Entry, len(entries))

		for pid, entry := range entries {
			state.Scores[key][fmt.Sprintf("%d", pid)] = entry
		}
	}

	for uniqueID, blob := range s.commonData {
		state.CommonData[fmt.Sprintf("%d", uniqueID)] = blob
	}

	contents, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode ranking state: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", filepath.Dir(s.path), err)
	}

	temporary := s.path + ".tmp"
	if err := os.WriteFile(temporary, contents, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", temporary, err)
	}

	if err := os.Rename(temporary, s.path); err != nil {
		return fmt.Errorf("failed to replace %s: %w", s.path, err)
	}

	return nil
}
















func (s *Store) InsertRankingByPIDAndRankingScoreData(pid types.PID, scoreData rankingtypes.RankingScoreData, uniqueID types.UInt64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	category := uint32(scoreData.Category)

	if _, ok := s.scores[category]; !ok {
		s.scores[category] = make(map[uint64]*Entry)
	}

	candidate := &Entry{
		PID:        uint64(pid),
		UniqueID:   uint64(uniqueID),
		Category:   category,
		Score:      uint32(scoreData.Score),
		OrderBy:    uint8(scoreData.OrderBy),
		Groups:     append([]byte(nil), scoreData.Groups...),
		Param:      uint64(scoreData.Param),
		UpdateTime: uint64(types.NewDateTime(0).Now()),
	}

	existing, hasExisting := s.scores[category][uint64(pid)]

	if hasExisting && scoreData.UpdateMode == rankingconstants.UpdateModeNormal {
		if !betterThan(candidate.Score, existing.Score, candidate.OrderBy) {

			return nil
		}
	}

	s.scores[category][uint64(pid)] = candidate

	return s.save()
}


func betterThan(score uint32, other uint32, orderBy uint8) bool {
	if rankingconstants.OrderBy(orderBy) == rankingconstants.OrderByDescending {
		return score > other
	}


	return score < other
}


func (s *Store) DeleteScore(pid uint64, category uint32) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if entries, ok := s.scores[category]; ok {
		delete(entries, pid)

		if len(entries) == 0 {
			delete(s.scores, category)
		}
	}

	return s.save()
}


func (s *Store) DeleteAllScores(pid uint64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for category, entries := range s.scores {
		delete(entries, pid)

		if len(entries) == 0 {
			delete(s.scores, category)
		}
	}

	return s.save()
}






func (s *Store) GetCommonData(uniqueID types.UInt64) (types.Buffer, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	blob, ok := s.commonData[uint64(uniqueID)]
	if !ok {
		return types.NewBuffer(nil), fmt.Errorf("no common data for unique ID %d", uint64(uniqueID))
	}

	return types.NewBuffer(append([]byte(nil), blob...)), nil
}


func (s *Store) UploadCommonData(pid types.PID, uniqueID types.UInt64, commonData types.Buffer) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	blob := append([]byte(nil), commonData...)
	wireUniqueID := uint64(uniqueID)
	s.commonData[perPlayerCommonDataKey(uint64(pid), wireUniqueID)] = blob




	if wireUniqueID == 0 {
		s.commonData[0] = append([]byte(nil), blob...)
	}

	return s.save()
}


func (s *Store) DeleteCommonData(uniqueID uint64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.commonData, uniqueID)

	return s.save()
}







func (s *Store) rankedEntries(category uint32) []*Entry {
	entries := make([]*Entry, 0, len(s.scores[category]))
	for _, entry := range s.scores[category] {
		entries = append(entries, entry)
	}



	orderBy := uint8(0)
	if len(entries) > 0 {
		orderBy = entries[0].OrderBy
	}

	sort.SliceStable(entries, func(a, b int) bool {
		if entries[a].Score != entries[b].Score {
			return betterThan(entries[a].Score, entries[b].Score, orderBy)
		}


		return entries[a].UpdateTime < entries[b].UpdateTime
	})

	return entries
}





func (s *Store) rankData(ordered []*Entry, selected []*Entry, orderCalculation rankingconstants.OrderCalculation) types.List[rankingtypes.RankingRankData] {
	ranks := make(map[uint64]uint32, len(ordered))



	standard := orderCalculation == rankingconstants.OrderCalculation113

	var previousScore uint32
	var previousRank uint32

	for index, entry := range ordered {
		rank := uint32(index) + 1

		if standard && index > 0 && entry.Score == previousScore {
			rank = previousRank
		}

		ranks[entry.PID] = rank
		previousScore = entry.Score
		previousRank = rank
	}

	list := types.NewList[rankingtypes.RankingRankData]()

	for _, entry := range selected {
		data := rankingtypes.NewRankingRankData()
		data.PrincipalID = types.NewPID(entry.PID)
		data.UniqueID = types.NewUInt64(entry.UniqueID)
		data.Order = types.NewUInt32(ranks[entry.PID])
		data.Category = types.NewUInt32(entry.Category)
		data.Score = types.NewUInt32(entry.Score)
		data.Groups = types.NewBuffer(append([]byte(nil), entry.Groups...))
		data.Param = types.NewUInt64(entry.Param)
		commonData, ok := s.commonData[perPlayerCommonDataKey(entry.PID, entry.UniqueID)]
		if !ok {


			commonData = s.commonData[entry.UniqueID]
		}
		data.CommonData = types.NewBuffer(append([]byte(nil), commonData...))
		data.UpdateTime = types.NewDateTime(entry.UpdateTime)

		list = append(list, data)
	}

	return list
}


func slice(entries []*Entry, orderParam rankingtypes.RankingOrderParam) []*Entry {
	offset := int(orderParam.Offset)
	length := int(orderParam.Length)

	if offset >= len(entries) {
		return nil
	}

	selected := entries[offset:]

	if length > 0 && length < len(selected) {
		selected = selected[:length]
	}

	return selected
}



func (s *Store) GetRankingsAndCountByCategoryAndRankingOrderParam(category types.UInt32, orderParam rankingtypes.RankingOrderParam) (types.List[rankingtypes.RankingRankData], uint32, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	ordered := s.rankedEntries(uint32(category))
	selected := slice(ordered, orderParam)

	return s.rankData(ordered, selected, orderParam.OrderCalculation), uint32(len(ordered)), nil
}



func (s *Store) GetNearbyRankingsAndCountByCategoryAndRankingOrderParam(pid types.PID, category types.UInt32, orderParam rankingtypes.RankingOrderParam) (types.List[rankingtypes.RankingRankData], uint32, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	ordered := s.rankedEntries(uint32(category))



	position := -1
	for index, entry := range ordered {
		if entry.PID == uint64(pid) {
			position = index
			break
		}
	}

	if position < 0 {
		return types.NewList[rankingtypes.RankingRankData](), uint32(len(ordered)), nil
	}

	length := int(orderParam.Length)
	if length <= 0 {
		length = 1
	}

	start := position - length/2
	if start < 0 {
		start = 0
	}

	end := start + length
	if end > len(ordered) {
		end = len(ordered)
	}

	return s.rankData(ordered, ordered[start:end], orderParam.OrderCalculation), uint32(len(ordered)), nil
}


func (s *Store) GetFriendsRankingsAndCountByCategoryAndRankingOrderParam(pid types.PID, category types.UInt32, orderParam rankingtypes.RankingOrderParam) (types.List[rankingtypes.RankingRankData], uint32, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	friends := s.friendSet(pid)
	ordered := filterEntries(s.rankedEntries(uint32(category)), friends)
	selected := slice(ordered, orderParam)

	return s.rankData(ordered, selected, orderParam.OrderCalculation), uint32(len(ordered)), nil
}



func (s *Store) GetNearbyFriendsRankingsAndCountByCategoryAndRankingOrderParam(pid types.PID, category types.UInt32, orderParam rankingtypes.RankingOrderParam) (types.List[rankingtypes.RankingRankData], uint32, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	friends := s.friendSet(pid)
	ordered := filterEntries(s.rankedEntries(uint32(category)), friends)
	selected := slice(ordered, orderParam)

	return s.rankData(ordered, selected, orderParam.OrderCalculation), uint32(len(ordered)), nil
}


func (s *Store) GetOwnRankingByCategoryAndRankingOrderParam(pid types.PID, category types.UInt32, orderParam rankingtypes.RankingOrderParam) (types.List[rankingtypes.RankingRankData], uint32, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	ordered := s.rankedEntries(uint32(category))

	var own []*Entry
	for _, entry := range ordered {
		if entry.PID == uint64(pid) {
			own = append(own, entry)
			break
		}
	}

	return s.rankData(ordered, own, orderParam.OrderCalculation), uint32(len(ordered)), nil
}



func (s *Store) RankingsForPIDs(pids []uint64, category uint32, orderParam rankingtypes.RankingOrderParam) (types.List[rankingtypes.RankingRankData], uint32) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	wanted := make(map[uint64]bool, len(pids))
	for _, pid := range pids {
		wanted[pid] = true
	}

	ordered := s.rankedEntries(category)
	selected := filterEntries(ordered, wanted)

	return s.rankData(ordered, selected, orderParam.OrderCalculation), uint32(len(ordered))
}




func (s *Store) ApproximateOrder(category uint32, score uint32) uint32 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	ordered := s.rankedEntries(category)

	orderBy := uint8(0)
	if len(ordered) > 0 {
		orderBy = ordered[0].OrderBy
	}


	order := uint32(1)
	for _, entry := range ordered {
		if betterThan(entry.Score, score, orderBy) {
			order++
		}
	}

	return order
}



func (s *Store) friendSet(pid types.PID) map[uint64]bool {
	set := map[uint64]bool{uint64(pid): true}

	if s.GetUserFriendPIDs == nil {
		return set
	}

	for _, friend := range s.GetUserFriendPIDs(uint32(pid)) {
		set[uint64(friend)] = true
	}

	return set
}

func filterEntries(entries []*Entry, allowed map[uint64]bool) []*Entry {
	filtered := make([]*Entry, 0, len(entries))
	for _, entry := range entries {
		if allowed[entry.PID] {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}
