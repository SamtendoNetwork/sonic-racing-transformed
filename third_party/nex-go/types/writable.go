package types


type Writable interface {
	StringLengthSize() int
	PIDSize() int
	UseStructureHeader() bool
	CopyNew() Writable
	Write(data []byte)
	WriteUInt8(value uint8)
	WriteUInt16LE(value uint16)
	WriteUInt32LE(value uint32)
	WriteUInt64LE(value uint64)
	WriteInt8(value int8)
	WriteInt16LE(value int16)
	WriteInt32LE(value int32)
	WriteInt64LE(value int64)
	WriteFloat32LE(value float32)
	WriteFloat64LE(value float64)
	WriteBool(value bool)
	Bytes() []byte
}
