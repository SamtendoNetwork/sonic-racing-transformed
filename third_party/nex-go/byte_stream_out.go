package nex

import (
	"github.com/PretendoNetwork/nex-go/v2/types"
	crunch "github.com/superwhiskers/crunch/v3"
)


type ByteStreamOut struct {
	*crunch.Buffer
	LibraryVersions *LibraryVersions
	Settings        *ByteStreamSettings
}


func (bso *ByteStreamOut) StringLengthSize() int {
	size := 2

	if bso.Settings != nil {
		size = bso.Settings.StringLengthSize
	}

	return size
}


func (bso *ByteStreamOut) PIDSize() int {
	size := 4

	if bso.Settings != nil {
		size = bso.Settings.PIDSize
	}

	return size
}


func (bso *ByteStreamOut) UseStructureHeader() bool {
	useStructureHeader := false

	if bso.Settings != nil {
		useStructureHeader = bso.Settings.UseStructureHeader
	}

	return useStructureHeader
}


func (bso *ByteStreamOut) CopyNew() types.Writable {
	return NewByteStreamOut(bso.LibraryVersions, bso.Settings)
}


func (bso *ByteStreamOut) Write(data []byte) {
	bso.Grow(int64(len(data)))
	bso.WriteBytesNext(data)
}


func (bso *ByteStreamOut) WriteUInt8(u8 uint8) {
	bso.Grow(1)
	bso.WriteByteNext(byte(u8))
}


func (bso *ByteStreamOut) WriteUInt16LE(u16 uint16) {
	bso.Grow(2)
	bso.WriteU16LENext([]uint16{u16})
}


func (bso *ByteStreamOut) WriteUInt32LE(u32 uint32) {
	bso.Grow(4)
	bso.WriteU32LENext([]uint32{u32})
}


func (bso *ByteStreamOut) WriteUInt64LE(u64 uint64) {
	bso.Grow(8)
	bso.WriteU64LENext([]uint64{u64})
}


func (bso *ByteStreamOut) WriteInt8(s8 int8) {
	bso.Grow(1)
	bso.WriteByteNext(byte(s8))
}


func (bso *ByteStreamOut) WriteInt16LE(s16 int16) {
	bso.Grow(2)
	bso.WriteU16LENext([]uint16{uint16(s16)})
}


func (bso *ByteStreamOut) WriteInt32LE(s32 int32) {
	bso.Grow(4)
	bso.WriteU32LENext([]uint32{uint32(s32)})
}


func (bso *ByteStreamOut) WriteInt64LE(s64 int64) {
	bso.Grow(8)
	bso.WriteU64LENext([]uint64{uint64(s64)})
}


func (bso *ByteStreamOut) WriteFloat32LE(f32 float32) {
	bso.Grow(4)
	bso.WriteF32LENext([]float32{f32})
}


func (bso *ByteStreamOut) WriteFloat64LE(f64 float64) {
	bso.Grow(8)
	bso.WriteF64LENext([]float64{f64})
}


func (bso *ByteStreamOut) WriteBool(b bool) {
	var bVar uint8
	if b {
		bVar = 1
	}

	bso.Grow(1)
	bso.WriteByteNext(byte(bVar))
}


func NewByteStreamOut(libraryVersions *LibraryVersions, settings *ByteStreamSettings) *ByteStreamOut {
	return &ByteStreamOut{
		Buffer:          crunch.NewBuffer(),
		LibraryVersions: libraryVersions,
		Settings:        settings,
	}
}
