package types


type Readable interface {
	StringLengthSize() int
	PIDSize() int
	UseStructureHeader() bool
	Remaining() uint64
	ReadRemaining() []byte
	Read(length uint64) ([]byte, error)
	ReadUInt8() (uint8, error)
	ReadUInt16LE() (uint16, error)
	ReadUInt32LE() (uint32, error)
	ReadUInt64LE() (uint64, error)
	ReadInt8() (int8, error)
	ReadInt16LE() (int16, error)
	ReadInt32LE() (int32, error)
	ReadInt64LE() (int64, error)
	ReadFloat32LE() (float32, error)
	ReadFloat64LE() (float64, error)
	ReadBool() (bool, error)
}
