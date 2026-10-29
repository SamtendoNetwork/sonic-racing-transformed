package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)


type Int16 int16


func (i16 Int16) WriteTo(writable Writable) {
	writable.WriteInt16LE(int16(i16))
}


func (i16 *Int16) ExtractFrom(readable Readable) error {
	value, err := readable.ReadInt16LE()
	if err != nil {
		return err
	}

	*i16 = Int16(value)
	return nil
}


func (i16 Int16) Copy() RVType {
	return NewInt16(int16(i16))
}


func (i16 Int16) Equals(o RVType) bool {
	other, ok := o.(Int16)
	if !ok {
		return false
	}

	return i16 == other
}



func (i16 Int16) CopyRef() RVTypePtr {
	copied := i16.Copy().(Int16)
	return &copied
}




func (i16 *Int16) Deref() RVType {
	return *i16
}


func (i16 Int16) String() string {
	return fmt.Sprintf("%d", i16)
}


func (i16 *Int16) Scan(value any) error {
	if value == nil {
		*i16 = Int16(0)
		return nil
	}

	switch v := value.(type) {
	case int64:
		*i16 = Int16(v)
	case []byte, string:
		var str string
		if b, ok := v.([]byte); ok {
			str = string(b)
		} else {
			str = v.(string)
		}

		parsed, err := strconv.ParseInt(str, 10, 16)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into Int16: %w", v, err)
		}
		*i16 = Int16(parsed)
	default:
		return fmt.Errorf("cannot scan %T into Int16", value)
	}

	return nil
}


func (i16 Int16) Value() (driver.Value, error) {
	return int64(i16), nil
}


func NewInt16(input int16) Int16 {
	i16 := Int16(input)
	return i16
}
