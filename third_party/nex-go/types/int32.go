package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)


type Int32 int32


func (i32 Int32) WriteTo(writable Writable) {
	writable.WriteInt32LE(int32(i32))
}


func (i32 *Int32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadInt32LE()
	if err != nil {
		return err
	}

	*i32 = Int32(value)
	return nil
}


func (i32 Int32) Copy() RVType {
	return NewInt32(int32(i32))
}


func (i32 Int32) Equals(o RVType) bool {
	other, ok := o.(Int32)
	if !ok {
		return false
	}

	return i32 == other
}



func (i32 Int32) CopyRef() RVTypePtr {
	copied := i32.Copy().(Int32)
	return &copied
}




func (i32 *Int32) Deref() RVType {
	return *i32
}


func (i32 Int32) String() string {
	return fmt.Sprintf("%d", i32)
}


func (i32 *Int32) Scan(value any) error {
	if value == nil {
		*i32 = Int32(0)
		return nil
	}

	switch v := value.(type) {
	case int64:
		*i32 = Int32(v)
	case []byte, string:
		var str string
		if b, ok := v.([]byte); ok {
			str = string(b)
		} else {
			str = v.(string)
		}

		parsed, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into Int32: %w", v, err)
		}
		*i32 = Int32(parsed)
	default:
		return fmt.Errorf("cannot scan %T into Int32", value)
	}

	return nil
}


func (i32 Int32) Value() (driver.Value, error) {
	return int64(i32), nil
}


func NewInt32(input int32) Int32 {
	i32 := Int32(input)
	return i32
}
