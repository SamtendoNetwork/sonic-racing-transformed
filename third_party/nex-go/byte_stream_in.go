package nex

import (
	"errors"

	crunch "github.com/superwhiskers/crunch/v3"
)


type ByteStreamIn struct {
	*crunch.Buffer
	LibraryVersions *LibraryVersions
	Settings        *ByteStreamSettings
}


func (bsi *ByteStreamIn) StringLengthSize() int {
	size := 2

	if bsi.Settings != nil {
		size = bsi.Settings.StringLengthSize
	}

	return size
}


func (bsi *ByteStreamIn) PIDSize() int {
	size := 4

	if bsi.Settings != nil {
		size = bsi.Settings.PIDSize
	}

	return size
}


func (bsi *ByteStreamIn) UseStructureHeader() bool {
	useStructureHeader := false

	if bsi.Settings != nil {
		useStructureHeader = bsi.Settings.UseStructureHeader
	}

	return useStructureHeader
}


func (bsi *ByteStreamIn) Remaining() uint64 {
	return uint64(len(bsi.Bytes()[bsi.ByteOffset():]))
}


func (bsi *ByteStreamIn) ReadRemaining() []byte {

	remaining, _ := bsi.Read(uint64(bsi.Remaining()))

	return remaining
}


func (bsi *ByteStreamIn) Read(length uint64) ([]byte, error) {
	if bsi.Remaining() < length {
		return []byte{}, errors.New("Read is OOB")
	}

	return bsi.ReadBytesNext(int64(length)), nil
}


func (bsi *ByteStreamIn) ReadUInt8() (uint8, error) {
	if bsi.Remaining() < 1 {
		return 0, errors.New("Not enough data to read uint8")
	}

	return uint8(bsi.ReadByteNext()), nil
}


func (bsi *ByteStreamIn) ReadUInt16LE() (uint16, error) {
	if bsi.Remaining() < 2 {
		return 0, errors.New("Not enough data to read uint16")
	}

	return bsi.ReadU16LENext(1)[0], nil
}


func (bsi *ByteStreamIn) ReadUInt32LE() (uint32, error) {
	if bsi.Remaining() < 4 {
		return 0, errors.New("Not enough data to read uint32")
	}

	return bsi.ReadU32LENext(1)[0], nil
}


func (bsi *ByteStreamIn) ReadUInt64LE() (uint64, error) {
	if bsi.Remaining() < 8 {
		return 0, errors.New("Not enough data to read uint64")
	}

	return bsi.ReadU64LENext(1)[0], nil
}


func (bsi *ByteStreamIn) ReadInt8() (int8, error) {
	if bsi.Remaining() < 1 {
		return 0, errors.New("Not enough data to read int8")
	}

	return int8(bsi.ReadByteNext()), nil
}


func (bsi *ByteStreamIn) ReadInt16LE() (int16, error) {
	if bsi.Remaining() < 2 {
		return 0, errors.New("Not enough data to read int16")
	}

	return int16(bsi.ReadU16LENext(1)[0]), nil
}


func (bsi *ByteStreamIn) ReadInt32LE() (int32, error) {
	if bsi.Remaining() < 4 {
		return 0, errors.New("Not enough data to read int32")
	}

	return int32(bsi.ReadU32LENext(1)[0]), nil
}


func (bsi *ByteStreamIn) ReadInt64LE() (int64, error) {
	if bsi.Remaining() < 8 {
		return 0, errors.New("Not enough data to read int64")
	}

	return int64(bsi.ReadU64LENext(1)[0]), nil
}


func (bsi *ByteStreamIn) ReadFloat32LE() (float32, error) {
	if bsi.Remaining() < 4 {
		return 0, errors.New("Not enough data to read float32")
	}

	return bsi.ReadF32LENext(1)[0], nil
}


func (bsi *ByteStreamIn) ReadFloat64LE() (float64, error) {
	if bsi.Remaining() < 8 {
		return 0, errors.New("Not enough data to read float64")
	}

	return bsi.ReadF64LENext(1)[0], nil
}


func (bsi *ByteStreamIn) ReadBool() (bool, error) {
	if bsi.Remaining() < 1 {
		return false, errors.New("Not enough data to read bool")
	}

	return bsi.ReadByteNext() == 1, nil
}


func NewByteStreamIn(data []byte, libraryVersions *LibraryVersions, settings *ByteStreamSettings) *ByteStreamIn {
	return &ByteStreamIn{
		Buffer:          crunch.NewBuffer(data),
		LibraryVersions: libraryVersions,
		Settings:        settings,
	}
}
