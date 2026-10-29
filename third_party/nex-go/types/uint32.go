package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)


type UInt32 uint32


func (u32 UInt32) WriteTo(writable Writable) {
	writable.WriteUInt32LE(uint32(u32))
}


func (u32 *UInt32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadUInt32LE()
	if err != nil {
		return err
	}

	*u32 = UInt32(value)
	return nil
}


func (u32 UInt32) Copy() RVType {
	return NewUInt32(uint32(u32))
}


func (u32 UInt32) Equals(o RVType) bool {
	other, ok := o.(UInt32)
	if !ok {
		return false
	}

	return u32 == other
}



func (u32 UInt32) CopyRef() RVTypePtr {
	copied := u32.Copy().(UInt32)
	return &copied
}




func (u32 *UInt32) Deref() RVType {
	return *u32
}


func (u32 UInt32) String() string {
	return fmt.Sprintf("%d", u32)
}


func (u32 *UInt32) Scan(value any) error {
	if value == nil {
		*u32 = UInt32(0)
		return nil
	}

	switch v := value.(type) {
	case int64:
		*u32 = UInt32(v)
	case []byte, string:
		var str string
		if b, ok := v.([]byte); ok {
			str = string(b)
		} else {
			str = v.(string)
		}

		parsed, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into UInt32: %w", v, err)
		}
		*u32 = UInt32(parsed)
	default:
		return fmt.Errorf("cannot scan %T into UInt32", value)
	}

	return nil
}


func (u32 UInt32) Value() (driver.Value, error) {
	return int64(u32), nil
}


func NewUInt32(input uint32) UInt32 {
	u32 := UInt32(input)
	return u32
}
