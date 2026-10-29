package nexserver

import (
	"encoding/binary"
	"sync"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/nex-server/sonic-transformed/globals"
)
























const (

	StructureHeaderAuto = "auto"

	StructureHeaderOn = "on"

	StructureHeaderOff = "off"
)

var (
	structureHeaderOnce   sync.Once
	structureHeaderMutex  sync.Mutex
	structureHeaderKnown  bool
	structureHeaderInUse  bool
	structureHeaderSource string
)


func applyStructureHeaderMode() {
	switch globals.Settings.StructureHeader {
	case StructureHeaderOn:
		setStructureHeader(true, "forced by SONIC_STRUCTURE_HEADER=on")
	case StructureHeaderOff:
		setStructureHeader(false, "forced by SONIC_STRUCTURE_HEADER=off")
	default:
		globals.Logger.Info("Structure headers: auto-detecting from the first login")
	}
}






func setStructureHeader(useHeader bool, reason string) {
	structureHeaderMutex.Lock()
	defer structureHeaderMutex.Unlock()

	structureHeaderKnown = true
	structureHeaderInUse = useHeader
	structureHeaderSource = reason

	if globals.AuthenticationServer != nil {
		globals.AuthenticationServer.ByteStreamSettings.UseStructureHeader = useHeader
	}

	if globals.SecureServer != nil {
		globals.SecureServer.ByteStreamSettings.UseStructureHeader = useHeader
	}

	state := "disabled"
	if useHeader {
		state = "enabled"
	}

	globals.Logger.Successf("Structure headers %s (%s)", state, reason)
}



















func detectStructureHeader(packet nex.PacketInterface) {
	if globals.Settings.StructureHeader != StructureHeaderAuto {
		return
	}

	structureHeaderMutex.Lock()
	decided := structureHeaderKnown
	structureHeaderMutex.Unlock()

	if decided {
		return
	}

	request := packet.RMCMessage()
	if request == nil || !request.IsRequest {
		return
	}

	objectData, ok := loginDataObject(request.ProtocolID, request.MethodID, request.Parameters)
	if !ok {
		return
	}

	useHeader, ok := objectHasStructureHeader(objectData)
	if !ok {
		globals.Logger.Warning("Could not determine structure header use from the first login; keeping the default")
		return
	}

	structureHeaderOnce.Do(func() {
		setStructureHeader(useHeader, "detected from the console's login data")
	})
}




func loginDataObject(protocolID uint16, methodID uint32, parameters []byte) ([]byte, bool) {
	cursor := 0

	switch {
	case protocolID == 0x0A && methodID == 0x02:

		var ok bool
		if cursor, ok = skipString(parameters, cursor); !ok {
			return nil, false
		}

	case protocolID == 0x0B && methodID == 0x04:

		var ok bool
		if cursor, ok = skipStationURLList(parameters, cursor); !ok {
			return nil, false
		}

	default:
		return nil, false
	}


	cursor, ok := skipString(parameters, cursor)
	if !ok {
		return nil, false
	}

	if cursor+8 > len(parameters) {
		return nil, false
	}

	cursor += 4

	objectLength := int(binary.LittleEndian.Uint32(parameters[cursor : cursor+4]))
	cursor += 4

	if objectLength < 0 || cursor+objectLength > len(parameters) {
		return nil, false
	}

	return parameters[cursor : cursor+objectLength], true
}















func objectHasStructureHeader(object []byte) (bool, bool) {
	if len(object) == 0 {
		return false, false
	}

	const maxStructureVersion = 8

	cursor := 0
	chain := 0

	for cursor+5 <= len(object) {
		version := object[cursor]
		length := int(binary.LittleEndian.Uint32(object[cursor+1 : cursor+5]))

		if version > maxStructureVersion || length < 0 || cursor+5+length > len(object) {
			break
		}

		cursor += 5 + length
		chain++
	}

	if chain > 0 && cursor == len(object) {
		return true, true
	}



	if len(object) >= 2 {
		stringLength := int(binary.LittleEndian.Uint16(object[0:2]))
		if stringLength > 0 && 2+stringLength <= len(object) {
			return false, true
		}
	}

	return false, false
}


func skipString(data []byte, cursor int) (int, bool) {
	if cursor+2 > len(data) {
		return cursor, false
	}

	length := int(binary.LittleEndian.Uint16(data[cursor : cursor+2]))
	cursor += 2

	if cursor+length > len(data) {
		return cursor, false
	}

	return cursor + length, true
}



func skipStationURLList(data []byte, cursor int) (int, bool) {
	if cursor+4 > len(data) {
		return cursor, false
	}

	count := int(binary.LittleEndian.Uint32(data[cursor : cursor+4]))
	cursor += 4


	if count < 0 || count > 64 {
		return cursor, false
	}

	for i := 0; i < count; i++ {
		var ok bool
		if cursor, ok = skipString(data, cursor); !ok {
			return cursor, false
		}
	}

	return cursor, true
}
