package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)


type UInt16 uint16


func (u16 UInt16) WriteTo(writable Writable) {
	writable.WriteUInt16LE(uint16(u16))
}


func (u16 *UInt16) ExtractFrom(readable Readable) error {
	value, err := readable.ReadUInt16LE()
	if err != nil {
		return err
	}

	*u16 = UInt16(value)
	return nil
}


func (u16 UInt16) Copy() RVType {
	return NewUInt16(uint16(u16))
}


func (u16 UInt16) Equals(o RVType) bool {
	other, ok := o.(UInt16)
	if !ok {
		return false
	}

	return u16 == other
}



func (u16 UInt16) CopyRef() RVTypePtr {
	copied := u16.Copy().(UInt16)
	return &copied
}




func (u16 *UInt16) Deref() RVType {
	return *u16
}


func (u16 UInt16) String() string {
	return fmt.Sprintf("%d", u16)
}


func (u16 *UInt16) Scan(value any) error {
	if value == nil {
		*u16 = UInt16(0)
		return nil
	}

	switch v := value.(type) {
	case int64:
		*u16 = UInt16(v)
	case []byte, string:
		var str string
		if b, ok := v.([]byte); ok {
			str = string(b)
		} else {
			str = v.(string)
		}

		parsed, err := strconv.ParseUint(str, 10, 16)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into UInt16: %w", v, err)
		}
		*u16 = UInt16(parsed)
	default:
		return fmt.Errorf("cannot scan %T into UInt16", value)
	}

	return nil
}


func (u16 UInt16) Value() (driver.Value, error) {
	return int64(u16), nil
}


func NewUInt16(input uint16) UInt16 {
	u16 := UInt16(input)
	return u16
}
