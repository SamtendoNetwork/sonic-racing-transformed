package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
)




type QBuffer []byte


func (qb QBuffer) WriteTo(writable Writable) {
	length := len(qb)

	writable.WriteUInt16LE(uint16(length))

	if length > 0 {
		writable.Write(qb)
	}
}


func (qb *QBuffer) ExtractFrom(readable Readable) error {
	length, err := readable.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read NEX qBuffer length. %s", err.Error())
	}

	data, err := readable.Read(uint64(length))
	if err != nil {
		return fmt.Errorf("Failed to read NEX qBuffer data. %s", err.Error())
	}

	*qb = data
	return nil
}


func (qb QBuffer) Copy() RVType {
	return NewQBuffer(qb)
}


func (qb QBuffer) Equals(o RVType) bool {
	if _, ok := o.(QBuffer); !ok {
		return false
	}

	return bytes.Equal(qb, o.(QBuffer))
}



func (qb QBuffer) CopyRef() RVTypePtr {
	copied := qb.Copy().(QBuffer)
	return &copied
}




func (qb *QBuffer) Deref() RVType {
	return *qb
}


func (qb QBuffer) String() string {
	return hex.EncodeToString(qb)
}


func (qb *QBuffer) Scan(value any) error {
	if value == nil {
		*qb = QBuffer([]byte{})
		return nil
	}

	if _, ok := value.([]byte); !ok {
		return fmt.Errorf("cannot scan %T into QBuffer", value)
	}

	*qb = QBuffer(value.([]byte))

	return nil
}


func (qb QBuffer) Value() (driver.Value, error) {
	return []byte(qb), nil
}


func NewQBuffer(input []byte) QBuffer {
	qb := make(QBuffer, len(input))
	copy(qb, input)

	return qb
}
