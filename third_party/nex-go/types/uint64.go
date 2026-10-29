package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)


type UInt64 uint64


func (u64 UInt64) WriteTo(writable Writable) {
	writable.WriteUInt64LE(uint64(u64))
}


func (u64 *UInt64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadUInt64LE()
	if err != nil {
		return err
	}

	*u64 = UInt64(value)
	return nil
}


func (u64 UInt64) Copy() RVType {
	return NewUInt64(uint64(u64))
}


func (u64 UInt64) Equals(o RVType) bool {
	other, ok := o.(UInt64)
	if !ok {
		return false
	}

	return u64 == other
}



func (u64 UInt64) CopyRef() RVTypePtr {
	copied := u64.Copy().(UInt64)
	return &copied
}




func (u64 *UInt64) Deref() RVType {
	return *u64
}


func (u64 UInt64) String() string {
	return fmt.Sprintf("%d", u64)
}


func (u64 *UInt64) Scan(value any) error {
	if value == nil {
		*u64 = UInt64(0)
		return nil
	}

	switch v := value.(type) {
	case int64:
		*u64 = UInt64(uint64(v))
	case []byte, string:
		var str string
		if b, ok := v.([]byte); ok {
			str = string(b)
		} else {
			str = v.(string)
		}

		parsed, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into UInt64: %w", v, err)
		}
		*u64 = UInt64(parsed)
	default:
		return fmt.Errorf("cannot scan %T into UInt64", value)
	}

	return nil
}


func (u64 UInt64) Value() (driver.Value, error) {
	return fmt.Sprintf("%d", uint64(u64)), nil
}


func NewUInt64(input uint64) UInt64 {
	u64 := UInt64(input)
	return u64
}
