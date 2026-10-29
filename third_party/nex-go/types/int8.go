package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)


type Int8 int8


func (i8 Int8) WriteTo(writable Writable) {
	writable.WriteInt8(int8(i8))
}


func (i8 *Int8) ExtractFrom(readable Readable) error {
	value, err := readable.ReadInt8()
	if err != nil {
		return err
	}

	*i8 = Int8(value)
	return nil
}


func (i8 Int8) Copy() RVType {
	return NewInt8(int8(i8))
}


func (i8 Int8) Equals(o RVType) bool {
	other, ok := o.(Int8)
	if !ok {
		return false
	}

	return i8 == other
}



func (i8 Int8) CopyRef() RVTypePtr {
	copied := i8.Copy().(Int8)
	return &copied
}




func (i8 *Int8) Deref() RVType {
	return *i8
}


func (i8 Int8) String() string {
	return fmt.Sprintf("%d", i8)
}


func (i8 *Int8) Scan(value any) error {
	if value == nil {
		*i8 = Int8(0)
		return nil
	}

	switch v := value.(type) {
	case int64:
		*i8 = Int8(v)
	case []byte, string:
		var str string
		if b, ok := v.([]byte); ok {
			str = string(b)
		} else {
			str = v.(string)
		}

		parsed, err := strconv.ParseInt(str, 10, 8)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into Int8: %w", v, err)
		}
		*i8 = Int8(parsed)
	default:
		return fmt.Errorf("cannot scan %T into Int8", value)
	}

	return nil
}


func (i8 Int8) Value() (driver.Value, error) {
	return int64(i8), nil
}


func NewInt8(input int8) Int8 {
	i8 := Int8(input)
	return i8
}
