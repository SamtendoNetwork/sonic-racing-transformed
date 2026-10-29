package nex

import (
	"errors"
	"fmt"

	"github.com/PretendoNetwork/nex-go/v2/types"
)


type RMCMessage struct {
	Endpoint         EndpointInterface
	IsRequest        bool
	IsSuccess        bool
	IsHPP            bool
	ProtocolID       uint16
	ProtocolName     types.String
	CallID           uint32
	MethodID         uint32
	MethodName       types.String
	ErrorCode        uint32
	VersionContainer *types.ClassVersionContainer
	Parameters       []byte

}


func (rmc *RMCMessage) Copy() *RMCMessage {
	copied := NewRMCMessage(rmc.Endpoint)

	copied.IsRequest = rmc.IsRequest
	copied.IsSuccess = rmc.IsSuccess
	copied.IsHPP = rmc.IsHPP
	copied.ProtocolID = rmc.ProtocolID
	copied.ProtocolName = rmc.ProtocolName
	copied.CallID = rmc.CallID
	copied.MethodID = rmc.MethodID
	copied.MethodName = rmc.MethodName
	copied.ErrorCode = rmc.ErrorCode

	if rmc.Parameters != nil {
		copied.Parameters = append([]byte(nil), rmc.Parameters...)
	}

	return copied
}


func (rmc *RMCMessage) FromBytes(data []byte) error {
	if rmc.Endpoint.UseVerboseRMC() {
		return rmc.decodeVerbose(data)
	} else {
		return rmc.decodePacked(data)
	}
}

func (rmc *RMCMessage) decodePacked(data []byte) error {
	stream := NewByteStreamIn(data, rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())

	length, err := stream.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read RMC Message size. %s", err.Error())
	}

	if stream.Remaining() != uint64(length) {
		return errors.New("RMC Message has unexpected size")
	}

	protocolID, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read RMC Message protocol ID. %s", err.Error())
	}

	rmc.ProtocolID = uint16(protocolID & ^byte(0x80))

	if rmc.ProtocolID == 0x7F {
		rmc.ProtocolID, err = stream.ReadUInt16LE()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message extended protocol ID. %s", err.Error())
		}
	}

	if protocolID&0x80 != 0 {
		rmc.IsRequest = true
		rmc.CallID, err = stream.ReadUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message (request) call ID. %s", err.Error())
		}

		rmc.MethodID, err = stream.ReadUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message (request) method ID. %s", err.Error())
		}

		rmc.Parameters = stream.ReadRemaining()
	} else {
		rmc.IsRequest = false
		rmc.IsSuccess, err = stream.ReadBool()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message (response) error check. %s", err.Error())
		}

		if rmc.IsSuccess {
			rmc.CallID, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) call ID. %s", err.Error())
			}

			rmc.MethodID, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) method ID. %s", err.Error())
			}

			rmc.MethodID = rmc.MethodID & ^uint32(0x8000)
			rmc.Parameters = stream.ReadRemaining()

		} else {
			rmc.ErrorCode, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) error code. %s", err.Error())
			}

			rmc.CallID, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) call ID. %s", err.Error())
			}

		}
	}

	return nil
}

func (rmc *RMCMessage) decodeVerbose(data []byte) error {
	stream := NewByteStreamIn(data, rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())

	length, err := stream.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read RMC Message size. %s", err.Error())
	}

	if stream.Remaining() != uint64(length) {
		return errors.New("RMC Message has unexpected size")
	}

	rmc.ProtocolName = types.NewString("")
	if err := rmc.ProtocolName.ExtractFrom(stream); err != nil {
		return fmt.Errorf("Failed to read RMC Message protocol name. %s", err.Error())
	}

	rmc.IsRequest, err = stream.ReadBool()
	if err != nil {
		return fmt.Errorf("Failed to read RMC Message \"is request\" bool. %s", err.Error())
	}

	if rmc.IsRequest {
		rmc.CallID, err = stream.ReadUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message (request) call ID. %s", err.Error())
		}

		rmc.MethodName = types.NewString("")
		if err := rmc.MethodName.ExtractFrom(stream); err != nil {
			return fmt.Errorf("Failed to read RMC Message (request) method name. %s", err.Error())
		}

		versionContainer := types.NewClassVersionContainer()
		if err := versionContainer.ExtractFrom(stream); err != nil {
			return fmt.Errorf("Failed to read RMC Message ClassVersionContainer. %s", err.Error())
		}

		rmc.VersionContainer = &versionContainer
		rmc.Parameters = stream.ReadRemaining()
	} else {
		rmc.IsSuccess, err = stream.ReadBool()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message (response) error check. %s", err.Error())
		}

		if rmc.IsSuccess {
			rmc.CallID, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) call ID. %s", err.Error())
			}

			rmc.MethodName = types.NewString("")
			if err := rmc.MethodName.ExtractFrom(stream); err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) method name. %s", err.Error())
			}

			rmc.Parameters = stream.ReadRemaining()

		} else {
			rmc.ErrorCode, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) error code. %s", err.Error())
			}

			rmc.CallID, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) call ID. %s", err.Error())
			}
		}
	}

	return nil
}


func (rmc *RMCMessage) Bytes() []byte {
	if rmc.Endpoint.UseVerboseRMC() {
		return rmc.encodeVerbose()
	} else {
		return rmc.encodePacked()
	}
}

func (rmc *RMCMessage) encodePacked() []byte {
	stream := NewByteStreamOut(rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())


	var protocolIDFlag uint16 = 0x80
	if !rmc.IsRequest {
		protocolIDFlag = 0
	}




	if !rmc.IsHPP || (rmc.IsHPP && rmc.IsRequest) {
		if rmc.ProtocolID < 0x80 {
			stream.WriteUInt8(uint8(rmc.ProtocolID | protocolIDFlag))
		} else {
			stream.WriteUInt8(uint8(0x7F | protocolIDFlag))
			stream.WriteUInt16LE(rmc.ProtocolID)
		}
	}

	if rmc.IsRequest {
		stream.WriteUInt32LE(rmc.CallID)
		stream.WriteUInt32LE(rmc.MethodID)

		if rmc.Parameters != nil && len(rmc.Parameters) > 0 {
			stream.Grow(int64(len(rmc.Parameters)))
			stream.WriteBytesNext(rmc.Parameters)
		}
	} else {
		stream.WriteBool(rmc.IsSuccess)

		if rmc.IsSuccess {
			stream.WriteUInt32LE(rmc.CallID)
			stream.WriteUInt32LE(rmc.MethodID | 0x8000)

			if rmc.Parameters != nil && len(rmc.Parameters) > 0 {
				stream.Grow(int64(len(rmc.Parameters)))
				stream.WriteBytesNext(rmc.Parameters)
			}
		} else {
			stream.WriteUInt32LE(uint32(rmc.ErrorCode))
			stream.WriteUInt32LE(rmc.CallID)
		}
	}

	serialized := stream.Bytes()

	message := NewByteStreamOut(rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())

	message.WriteUInt32LE(uint32(len(serialized)))
	message.Grow(int64(len(serialized)))
	message.WriteBytesNext(serialized)

	return message.Bytes()
}

func (rmc *RMCMessage) encodeVerbose() []byte {
	stream := NewByteStreamOut(rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())

	rmc.ProtocolName.WriteTo(stream)
	stream.WriteBool(rmc.IsRequest)

	if rmc.IsRequest {
		stream.WriteUInt32LE(rmc.CallID)
		rmc.MethodName.WriteTo(stream)

		if rmc.VersionContainer != nil {
			rmc.VersionContainer.WriteTo(stream)
		} else {

			stream.WriteUInt32LE(0)
		}

		if rmc.Parameters != nil && len(rmc.Parameters) > 0 {
			stream.Grow(int64(len(rmc.Parameters)))
			stream.WriteBytesNext(rmc.Parameters)
		}
	} else {
		stream.WriteBool(rmc.IsSuccess)

		if rmc.IsSuccess {
			stream.WriteUInt32LE(rmc.CallID)
			rmc.MethodName.WriteTo(stream)

			if rmc.Parameters != nil && len(rmc.Parameters) > 0 {
				stream.Grow(int64(len(rmc.Parameters)))
				stream.WriteBytesNext(rmc.Parameters)
			}
		} else {
			stream.WriteUInt32LE(uint32(rmc.ErrorCode))
			stream.WriteUInt32LE(rmc.CallID)
		}
	}

	serialized := stream.Bytes()

	message := NewByteStreamOut(rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())

	message.WriteUInt32LE(uint32(len(serialized)))
	message.Grow(int64(len(serialized)))
	message.WriteBytesNext(serialized)

	return message.Bytes()
}


func NewRMCMessage(endpoint EndpointInterface) *RMCMessage {
	return &RMCMessage{
		Endpoint: endpoint,
	}
}


func NewRMCRequest(endpoint EndpointInterface) *RMCMessage {
	return &RMCMessage{
		Endpoint:  endpoint,
		IsRequest: true,
	}
}


func NewRMCSuccess(endpoint EndpointInterface, parameters []byte) *RMCMessage {
	message := NewRMCMessage(endpoint)
	message.IsRequest = false
	message.IsSuccess = true
	message.Parameters = parameters

	return message
}


func NewRMCError(endpoint EndpointInterface, errorCode uint32) *RMCMessage {
	if int(errorCode)&errorMask == 0 {
		errorCode = uint32(int(errorCode) | errorMask)
	}

	message := NewRMCMessage(endpoint)
	message.IsRequest = false
	message.IsSuccess = false
	message.ErrorCode = errorCode

	return message
}
