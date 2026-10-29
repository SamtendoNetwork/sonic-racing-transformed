package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
)




type Buffer []byte


func (b Buffer) WriteTo(writable Writable) {
	length := len(b)

	writable.WriteUInt32LE(uint32(length))

	if length > 0 {
		writable.Write(b)
	}
}


func (b *Buffer) ExtractFrom(readable Readable) error {
	length, err := readable.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read NEX Buffer length. %s", err.Error())
	}

	value, err := readable.Read(uint64(length))
	if err != nil {
		return fmt.Errorf("Failed to read NEX Buffer data. %s", err.Error())
	}

	*b = Buffer(value)
	return nil
}


func (b Buffer) Copy() RVType {
	return NewBuffer(b)
}


func (b Buffer) Equals(o RVType) bool {
	if _, ok := o.(Buffer); !ok {
		return false
	}

	return bytes.Equal(b, o.(Buffer))
}



func (b Buffer) CopyRef() RVTypePtr {
	copied := b.Copy().(Buffer)
	return &copied
}




func (b *Buffer) Deref() RVType {
	return *b
}


func (b Buffer) String() string {
	return hex.EncodeToString(b)
}


func (b *Buffer) Scan(value any) error {
	if value == nil {
		*b = Buffer([]byte{})
		return nil
	}

	if _, ok := value.([]byte); !ok {
		return fmt.Errorf("cannot scan %T into Buffer", value)
	}

	*b = Buffer(value.([]byte))

	return nil
}


func (b Buffer) Value() (driver.Value, error) {
	return []byte(b), nil
}


func NewBuffer(input []byte) Buffer {
	b := make(Buffer, len(input))
	copy(b, input)

	return b
}
