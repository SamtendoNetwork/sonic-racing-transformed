package database

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	"github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making/tracking"
	match_making_constants "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/constants"
	match_making_types "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
	notifications_constants "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/constants"
	notifications_types "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/types"
	pqextended "github.com/PretendoNetwork/pq-extended"
)


func DisconnectParticipant(manager *common_globals.MatchmakingManager, connection *nex.PRUDPConnection) {
	var nexError *nex.Error

	rows, err := manager.Database.Query(`SELECT id, owner_pid, host_pid, min_participants, max_participants, participation_policy, policy_argument, flags, state, description, type, participants FROM matchmaking.gatherings WHERE $1 = ANY(participants) AND registered=true`, connection.PID())
	if err != nil {
		common_globals.Logger.Critical(err.Error())
		return
	}

	for rows.Next() {
		gathering := match_making_types.NewGathering()
		var gatheringType string
		var participants []uint64

		err = rows.Scan(
			&gathering.ID,
			&gathering.OwnerPID,
			&gathering.HostPID,
			&gathering.MinimumParticipants,
			&gathering.MaximumParticipants,
			&gathering.ParticipationPolicy,
			&gathering.PolicyArgument,
			&gathering.Flags,
			&gathering.State,
			&gathering.Description,
			&gatheringType,
			pqextended.Array(&participants),
		)
		if err != nil {
			common_globals.Logger.Critical(err.Error())
			continue
		}




		if gathering.Flags.HasFlag(match_making_constants.GatheringFlagPersistentGathering) || (gatheringType == "PersistentGathering" && !gathering.Flags.HasFlag(match_making_constants.GatheringFlagNoPersistentParticipation)) {
			continue
		}


		participants, nexError = RemoveParticipantFromGathering(manager, uint32(gathering.ID), uint64(connection.PID()))
		if nexError != nil {
			common_globals.Logger.Error(nexError.Error())
			continue
		}

		nexError = tracking.LogDisconnectGathering(manager.Database, connection.PID(), uint32(gathering.ID), participants)
		if nexError != nil {
			common_globals.Logger.Error(nexError.Error())
			continue
		}


		if gathering.Flags.HasFlag(match_making_constants.GatheringFlagPersistentGathering) || gathering.Flags.HasFlag(match_making_constants.GatheringFlagAllowNoParticipant) {
			continue
		}

		if len(participants) == 0 {


			nexError = UnregisterGathering(manager, connection.PID(), uint32(gathering.ID))
			if nexError != nil {
				common_globals.Logger.Error(nexError.Error())
			}

			continue
		}

		ownerPID := uint64(gathering.OwnerPID)
		if connection.PID().Equals(gathering.OwnerPID) {



			if !gathering.Flags.HasFlag(match_making_constants.GatheringFlagMigrateOwner) {
				nexError = UnregisterGathering(manager, connection.PID(), uint32(gathering.ID))
				if nexError != nil {
					common_globals.Logger.Error(nexError.Error())
					continue
				}

				oEvent := notifications_types.NewNotificationEvent()
				oEvent.PIDSource = connection.PID().Copy().(types.PID)
				oEvent.Type = notifications_constants.NotificationCategoryGatheringUnregistered.Build()
				oEvent.Param1 = types.UInt64(gathering.ID)

				common_globals.SendNotificationEvent(connection.Endpoint().(*nex.PRUDPEndPoint), oEvent, common_globals.RemoveDuplicates(participants))

				continue
			}

			ownerPID, nexError = MigrateGatheringOwnership(manager, connection, gathering, participants)
			if nexError != nil {
				common_globals.Logger.Error(nexError.Error())
			}
		}



		if connection.PID().Equals(gathering.HostPID) && ownerPID != 0 {
			nexError = UpdateSessionHost(manager, uint32(gathering.ID), types.NewPID(ownerPID), types.NewPID(ownerPID))
			if nexError != nil {
				common_globals.Logger.Error(nexError.Error())
			} else {
				oEvent := notifications_types.NewNotificationEvent()
				oEvent.PIDSource = connection.PID().Copy().(types.PID)
				oEvent.Type = notifications_constants.NotificationCategoryHostChangeEvent.Build()
				oEvent.Param1 = types.UInt64(gathering.ID)


				common_globals.SendNotificationEvent(connection.Endpoint().(*nex.PRUDPEndPoint), oEvent, common_globals.RemoveDuplicates(participants))

				nexError = tracking.LogChangeHost(manager.Database, connection.PID(), uint32(gathering.ID), gathering.HostPID, types.NewPID(ownerPID))
				if nexError != nil {
					common_globals.Logger.Error(nexError.Error())
					continue
				}
			}
		}

		oEvent := notifications_types.NewNotificationEvent()
		oEvent.PIDSource = connection.PID().Copy().(types.PID)
		oEvent.Type = notifications_constants.NotificationCategoryParticipationEvent.Build(notifications_constants.ParticipationEventsDisconnect)
		oEvent.Param1 = types.UInt64(gathering.ID)
		oEvent.Param2 = types.UInt64(connection.PID())

		var participantEndedTargets []uint64


		if gathering.Flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipants) || gathering.Flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipantsReproducibly) {
			participantEndedTargets = common_globals.RemoveDuplicates(participants)
		} else {
			participantEndedTargets = []uint64{uint64(gathering.OwnerPID)}
		}


		common_globals.SendNotificationEvent(connection.Endpoint().(*nex.PRUDPEndPoint), oEvent, participantEndedTargets)
	}

	rows.Close()
}
