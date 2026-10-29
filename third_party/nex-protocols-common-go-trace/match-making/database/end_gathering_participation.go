package database

import (
	"slices"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	"github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making/tracking"
	match_making_constants "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/constants"
	notifications_constants "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/constants"
	notifications_types "github.com/PretendoNetwork/nex-protocols-go/v2/notifications/types"
)


func EndGatheringParticipation(manager *common_globals.MatchmakingManager, gatheringID uint32, connection *nex.PRUDPConnection, message string) *nex.Error {
	gathering, _, participants, _, nexError := FindGatheringByID(manager, gatheringID)
	if nexError != nil {
		return nexError
	}


	if !slices.Contains(participants, uint64(connection.PID())) {
		return nex.NewError(nex.ResultCodes.RendezVous.NotParticipatedGathering, "change_error")
	}

	newParticipants, nexError := RemoveParticipantFromGathering(manager, gatheringID, uint64(connection.PID()))
	if nexError != nil {
		return nexError
	}

	nexError = tracking.LogLeaveGathering(manager.Database, connection.PID(), gatheringID, newParticipants)
	if nexError != nil {
		return nexError
	}


	if gathering.Flags.HasFlag(match_making_constants.GatheringFlagPersistentGathering) || gathering.Flags.HasFlag(match_making_constants.GatheringFlagAllowNoParticipant) {
		return nil
	}

	if len(newParticipants) == 0 {

		return UnregisterGathering(manager, connection.PID(), gatheringID)
	}

	var ownerPID uint64 = uint64(gathering.OwnerPID)
	if connection.PID().Equals(gathering.OwnerPID) {



		if !gathering.Flags.HasFlag(match_making_constants.GatheringFlagMigrateOwner) {
			nexError = UnregisterGathering(manager, connection.PID(), gatheringID)
			if nexError != nil {
				return nexError
			}

			oEvent := notifications_types.NewNotificationEvent()
			oEvent.PIDSource = connection.PID().Copy().(types.PID)
			oEvent.Type = notifications_constants.NotificationCategoryGatheringUnregistered.Build()
			oEvent.Param1 = types.UInt64(gatheringID)

			common_globals.SendNotificationEvent(connection.Endpoint().(*nex.PRUDPEndPoint), oEvent, common_globals.RemoveDuplicates(newParticipants))

			return nil
		}

		ownerPID, nexError = MigrateGatheringOwnership(manager, connection, gathering, newParticipants)
		if nexError != nil {
			return nexError
		}
	}



	if connection.PID().Equals(gathering.HostPID) && ownerPID != 0 {
		nexError = UpdateSessionHost(manager, gatheringID, types.NewPID(ownerPID), types.NewPID(ownerPID))
		if nexError != nil {
			common_globals.Logger.Error(nexError.Error())
			return nexError
		}

		oEvent := notifications_types.NewNotificationEvent()
		oEvent.PIDSource = connection.PID().Copy().(types.PID)
		oEvent.Type = notifications_constants.NotificationCategoryHostChangeEvent.Build()
		oEvent.Param1 = types.UInt64(gatheringID)


		common_globals.SendNotificationEvent(connection.Endpoint().(*nex.PRUDPEndPoint), oEvent, common_globals.RemoveDuplicates(participants))

		nexError = tracking.LogChangeHost(manager.Database, connection.PID(), gatheringID, gathering.HostPID, types.NewPID(ownerPID))
		if nexError != nil {
			common_globals.Logger.Error(nexError.Error())
			return nexError
		}
	}

	oEvent := notifications_types.NewNotificationEvent()
	oEvent.PIDSource = connection.PID().Copy().(types.PID)
	oEvent.Type = notifications_constants.NotificationCategoryParticipationEvent.Build(notifications_constants.ParticipationEventsEndParticipation)
	oEvent.Param1 = types.UInt64(gatheringID)
	oEvent.Param2 = types.UInt64(connection.PID())
	oEvent.StrParam = types.NewString(message)

	var participationEndedTargets []uint64


	if gathering.Flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipants) || gathering.Flags.HasFlag(match_making_constants.GatheringFlagNotifyParticipationEventsToAllParticipantsReproducibly) {
		participationEndedTargets = common_globals.RemoveDuplicates(participants)
	} else {
		participationEndedTargets = []uint64{uint64(gathering.OwnerPID)}
	}

	common_globals.SendNotificationEvent(connection.Endpoint().(*nex.PRUDPEndPoint), oEvent, participationEndedTargets)

	return nil
}
