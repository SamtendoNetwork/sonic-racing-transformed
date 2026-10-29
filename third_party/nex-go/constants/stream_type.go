package constants








type StreamType uint8


func (st StreamType) EnumIndex() uint8 {
	return uint8(st)
}

const (

	StreamTypeDO StreamType = iota + 1


	StreamTypeRV


	StreamTypeOldRVSec


	StreamTypeSBMGMT


	StreamTypeNAT


	StreamTypeSessionDiscovery


	StreamTypeNATEcho


	StreamTypeRouting


	StreamTypeGame


	StreamTypeRVSecure


	StreamTypeRelay
)
