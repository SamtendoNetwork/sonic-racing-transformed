package database

import (
	"database/sql"
	"slices"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	"github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making/tracking"
	match_making_constants "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/constants"
	notifications_constants "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/constants"
	notifications_types "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/types"
	pqextended "github.com/PretendoNetwork/pq-extended"
)


func JoinGathering(manager *common_globals.MatchmakingManager, gatheringID uint32, connection *nex.PRUDPConnection, vacantParticipants uint16, joinMessage string) (uint32, *nex.Error) {


	if vacantParticipants == 0 {
		vacantParticipants = 1
	}

	var ownerPID uint64
	var maxParticipants uint32
	var flags match_making_constants.GatheringFlags
	var participants []uint64
	err := manager.Database.QueryRow(`SELECT owner_pid, max_participants, flags, participants FROM matchmaking.gatherings WHERE id=$1`, gatheringID).Scan(&ownerPID, &maxParticipants, &flags, pqextended.Array(&participants))
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nex.NewError(nex.ResultCodes.RendezVous.SessionVoid, "change_error")
		} else {
			return 0, nex.NewError(nex.ResultCodes.Core.Unknown, err.Error())
		}
	}

	if maxParticipants != 0 {
		if uint32(len(participants))+uint32(vacantParticipants) > maxParticipants {
			return 0, nex.NewError(nex.ResultCodes.RendezVous.SessionFull, "change_error")
		}
	}

	if slices.Contains(participants, uint64(connection.PID())) {
		return 0, nex.NewError(nex.ResultCodes.RendezVous.AlreadyParticipatedGathering, "change_error")
	}

	var newParticipants []uint64


	for range vacantParticipants {
		newParticipants = append(newParticipants, uint64(connection.PID()))
	}

	var totalParticipants []uint64 = append(newParticipants, participants...)


	_, err = manager.Database.Exec(`UPDATE matchmaking.gatherings SET participants=$1 WHERE id=$2`, pqextended.Array(totalParticipants), gatheringID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nex.NewError(nex.ResultCodes.RendezVous.SessionVoid, "change_error")
		} else {
			return 0, nex.NewError(nex.ResultCodes.Core.Unknown, err.Error())
		}
	}

	nexError := tracking.LogJoinGathering(manager.Database, connection.PID(), gatheringID, newParticipants, totalParticipants)
	if nexError != nil {
		return 0, nexError
	}

	var participantJoinedTargets []uint64


	if flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipants) || flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipantsReproducibly) {
		participantJoinedTargets = common_globals.RemoveDuplicates(totalParticipants)
	} else {


		if uint64(connection.PID()) == ownerPID {
			return uint32(len(totalParticipants)), nil
		}

		participantJoinedTargets = []uint64{ownerPID}
	}

	oEvent := notifications_types.NewNotificationEvent()
	oEvent.PIDSource = connection.PID()
	oEvent.Type = notifications_constants.NotificationCategoryParticipationEvent.Build(notifications_constants.ParticipationEventsParticipate)
	oEvent.Param1 = types.UInt64(gatheringID)
	oEvent.Param2 = types.UInt64(connection.PID())
	oEvent.StrParam = types.NewString(joinMessage)
	oEvent.Param3 = types.UInt64(len(totalParticipants))

	common_globals.SendNotificationEvent(connection.Endpoint().(*nex.PRUDPEndPoint), oEvent, participantJoinedTargets)


	if flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipantsReproducibly) {

		for _, participant := range common_globals.RemoveDuplicates(participants) {
			oEvent := notifications_types.NewNotificationEvent()
			oEvent.PIDSource = connection.PID()
			oEvent.Type = notifications_constants.NotificationCategoryParticipationEvent.Build(notifications_constants.ParticipationEventsParticipate)
			oEvent.Param1 = types.UInt64(gatheringID)
			oEvent.Param2 = types.UInt64(participant)
			oEvent.StrParam = types.NewString(joinMessage)
			oEvent.Param3 = types.UInt64(len(totalParticipants))


			common_globals.SendNotificationEvent(connection.Endpoint().(*nex.PRUDPEndPoint), oEvent, []uint64{uint64(connection.PID())})
		}
	}

	return uint32(len(totalParticipants)), nil
}
