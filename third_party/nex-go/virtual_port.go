package nex

import "github.com/PretendoNetwork/nex-go/v2/constants"







type VirtualPort byte


func (vp *VirtualPort) SetStreamType(streamType constants.StreamType) {
	*vp = VirtualPort((byte(*vp) & 0x0F) | (byte(streamType) << 4))
}


func (vp VirtualPort) StreamType() constants.StreamType {
	return constants.StreamType(vp >> 4)
}


func (vp *VirtualPort) SetStreamID(streamID uint8) {
	*vp = VirtualPort((byte(*vp) & 0xF0) | (streamID & 0x0F))
}


func (vp VirtualPort) StreamID() uint8 {
	return uint8(vp & 0xF)
}
