package datastore

import (
	"github.com/PretendoNetwork/nex-go/v2"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	datastore "github.com/PretendoNetwork/nex-protocols-go/v2/datastore"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
)

func (commonProtocol *CommonProtocol) postMetaBinary(err error, packet nex.PacketInterface, callID uint32, param datastore_types.DataStorePreparePostParam) (*nex.RMCMessage, *nex.Error) {







	if commonProtocol.InitializeObjectByPreparePostParam == nil {
		common_globals.Logger.Warning("InitializeObjectByPreparePostParam not defined")
		return nil, nex.NewError(nex.ResultCodes.Core.NotImplemented, "change_error")
	}

	if commonProtocol.InitializeObjectRatingWithSlot == nil {
		common_globals.Logger.Warning("InitializeObjectRatingWithSlot not defined")
		return nil, nex.NewError(nex.ResultCodes.Core.NotImplemented, "change_error")
	}

	if err != nil {
		common_globals.Logger.Error(err.Error())
		return nil, nex.NewError(nex.ResultCodes.DataStore.Unknown, "change_error")
	}

	connection := packet.Sender()
	endpoint := connection.Endpoint()


	dataID, errCode := commonProtocol.InitializeObjectByPreparePostParam(connection.PID(), param)
	if errCode != nil {
		common_globals.Logger.Errorf("Error code on object init: %s", errCode.Error())
		return nil, errCode
	}


	for _ , ratingInitParamWithSlot := range param.RatingInitParams {
		errCode = commonProtocol.InitializeObjectRatingWithSlot(dataID, ratingInitParamWithSlot)
		if errCode != nil {
			common_globals.Logger.Errorf("Error code on rating init: %s", errCode.Error())
			break
		}
	}

	if errCode != nil {
		return nil, errCode
	}

	rmcResponseStream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())

	rmcResponseStream.WriteUInt64LE(dataID)

	rmcResponseBody := rmcResponseStream.Bytes()

	rmcResponse := nex.NewRMCSuccess(endpoint, rmcResponseBody)
	rmcResponse.ProtocolID = datastore.ProtocolID
	rmcResponse.MethodID = datastore.MethodPostMetaBinary
	rmcResponse.CallID = callID

	if commonProtocol.OnAfterPostMetaBinary != nil {
		go commonProtocol.OnAfterPostMetaBinary(packet, param)
	}

	return rmcResponse, nil
}
