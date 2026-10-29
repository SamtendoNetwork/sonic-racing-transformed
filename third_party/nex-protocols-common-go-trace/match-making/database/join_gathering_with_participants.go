package database

import (
	"database/sql"
	"slices"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	"github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making/tracking"
	"github.com/PretendoNetwork/nex-protocols-go/v2/match-making/constants"
	match_making_constants "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/constants"
	notifications_constants "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/constants"
	notifications_types "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/types"
	pqextended "github.com/PretendoNetwork/pq-extended"
)


func JoinGatheringWithParticipants(manager *common_globals.MatchmakingManager, gatheringID uint32, connection *nex.PRUDPConnection, additionalParticipants []types.PID, joinMessage string, joinMatchmakeSessionBehavior constants.JoinMatchmakeSessionBehavior) (uint32, *nex.Error) {
	var ownerPID uint64
	var maxParticipants uint32
	var flags match_making_constants.GatheringFlags
	var oldParticipants []uint64
	err := manager.Database.QueryRow(`SELECT owner_pid, max_participants, flags, participants FROM matchmaking.gatherings WHERE id=$1`, gatheringID).Scan(&ownerPID, &maxParticipants, &flags, pqextended.Array(&oldParticipants))
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nex.NewError(nex.ResultCodes.RendezVous.SessionVoid, "change_error")
		} else {
			return 0, nex.NewError(nex.ResultCodes.Core.Unknown, err.Error())
		}
	}

	if uint32(len(oldParticipants)+1+len(additionalParticipants)) > maxParticipants {
		return 0, nex.NewError(nex.ResultCodes.RendezVous.SessionFull, "change_error")
	}

	var newParticipants []uint64


	if joinMatchmakeSessionBehavior == constants.JoinMatchmakeSessionBehaviorImAlreadyJoined {
		if !slices.Contains(oldParticipants, uint64(connection.PID())) {
			return 0, nex.NewError(nex.ResultCodes.RendezVous.NotParticipatedGathering, "change_error")
		}
	} else {
		if slices.Contains(oldParticipants, uint64(connection.PID())) {
			return 0, nex.NewError(nex.ResultCodes.RendezVous.AlreadyParticipatedGathering, "change_error")
		}


		newParticipants = []uint64{uint64(connection.PID())}
	}

	for _, participant := range additionalParticipants {
		newParticipants = append(newParticipants, uint64(participant))
	}

	participants := append(oldParticipants, newParticipants...)


	_, err = manager.Database.Exec(`UPDATE matchmaking.gatherings SET participants=$1 WHERE id=$2`, pqextended.Array(participants), gatheringID)
	if err != nil {
		return 0, nex.NewError(nex.ResultCodes.Core.Unknown, err.Error())
	}


	nexError := tracking.LogJoinGathering(manager.Database, connection.PID(), gatheringID, newParticipants, participants)
	if nexError != nil {
		return 0, nexError
	}

	var participantJoinedTargets []uint64


	if flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipants) || flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipantsReproducibly) {
		participantJoinedTargets = common_globals.RemoveDuplicates(participants)
	} else {
		participantJoinedTargets = []uint64{ownerPID}
	}


	for _, participant := range common_globals.RemoveDuplicates(newParticipants) {

		if uint64(connection.PID()) == participant {
			continue
		}

		oEvent := notifications_types.NewNotificationEvent()
		oEvent.PIDSource = connection.PID()
		oEvent.Type = notifications_constants.NotificationCategoryAddedToGathering.Build()
		oEvent.Param1 = types.UInt64(gatheringID)
		oEvent.Param2 = types.UInt64(participant)


		common_globals.SendNotificationEvent(connection.Endpoint().(*nex.PRUDPEndPoint), oEvent, []uint64{participant})
	}

	for _, participant := range newParticipants {


		if (flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipants) || flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipantsReproducibly)) || uint64(connection.PID()) != ownerPID {
			oEvent := notifications_types.NewNotificationEvent()
			oEvent.PIDSource = connection.PID()
			oEvent.Type = notifications_constants.NotificationCategoryParticipationEvent.Build(notifications_constants.ParticipationEventsParticipate)
			oEvent.Param1 = types.UInt64(gatheringID)
			oEvent.Param2 = types.UInt64(participant)
			oEvent.StrParam = types.NewString(joinMessage)
			oEvent.Param3 = types.UInt64(len(participants))

			common_globals.SendNotificationEvent(connection.Endpoint().(*nex.PRUDPEndPoint), oEvent, participantJoinedTargets)
		}


		if flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipantsReproducibly) {

			for _, oldParticipant := range common_globals.RemoveDuplicates(oldParticipants) {
				oEvent := notifications_types.NewNotificationEvent()
				oEvent.PIDSource = connection.PID()
				oEvent.Type = notifications_constants.NotificationCategoryParticipationEvent.Build(notifications_constants.ParticipationEventsParticipate)
				oEvent.Param1 = types.UInt64(gatheringID)
				oEvent.Param2 = types.UInt64(oldParticipant)
				oEvent.StrParam = types.NewString(joinMessage)
				oEvent.Param3 = types.UInt64(len(participants))


				common_globals.SendNotificationEvent(connection.Endpoint().(*nex.PRUDPEndPoint), oEvent, []uint64{participant})
			}
		}
	}

	return uint32(len(participants)), nil
}
