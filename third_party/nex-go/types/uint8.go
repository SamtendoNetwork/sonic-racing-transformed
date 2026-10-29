package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)


type UInt8 uint8


func (u8 UInt8) WriteTo(writable Writable) {
	writable.WriteUInt8(uint8(u8))
}


func (u8 *UInt8) ExtractFrom(readable Readable) error {
	value, err := readable.ReadUInt8()
	if err != nil {
		return err
	}

	*u8 = UInt8(value)
	return nil
}


func (u8 UInt8) Copy() RVType {
	return NewUInt8(uint8(u8))
}


func (u8 UInt8) Equals(o RVType) bool {
	other, ok := o.(UInt8)
	if !ok {
		return false
	}

	return u8 == other
}



func (u8 UInt8) CopyRef() RVTypePtr {
	copied := u8.Copy().(UInt8)
	return &copied
}




func (u8 *UInt8) Deref() RVType {
	return *u8
}


func (u8 UInt8) String() string {
	return fmt.Sprintf("%d", u8)
}


func (u8 *UInt8) Scan(value any) error {
	if value == nil {
		*u8 = UInt8(0)
		return nil
	}

	switch v := value.(type) {
	case int64:
		*u8 = UInt8(v)
	case []byte, string:
		var str string
		if b, ok := v.([]byte); ok {
			str = string(b)
		} else {
			str = v.(string)
		}

		parsed, err := strconv.ParseUint(str, 10, 8)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into UInt8: %w", v, err)
		}
		*u8 = UInt8(parsed)
	default:
		return fmt.Errorf("cannot scan %T into UInt8", value)
	}

	return nil
}


func (u8 UInt8) Value() (driver.Value, error) {
	return int64(u8), nil
}


func NewUInt8(input uint8) UInt8 {
	u8 := UInt8(input)
	return u8
}
