package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)


type Int64 int64


func (i64 Int64) WriteTo(writable Writable) {
	writable.WriteInt64LE(int64(i64))
}


func (i64 *Int64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadInt64LE()
	if err != nil {
		return err
	}

	*i64 = Int64(value)
	return nil
}


func (i64 Int64) Copy() RVType {
	return NewInt64(int64(i64))
}


func (i64 Int64) Equals(o RVType) bool {
	other, ok := o.(Int64)
	if !ok {
		return false
	}

	return i64 == other
}



func (i64 Int64) CopyRef() RVTypePtr {
	copied := i64.Copy().(Int64)
	return &copied
}




func (i64 *Int64) Deref() RVType {
	return *i64
}


func (i64 Int64) String() string {
	return fmt.Sprintf("%d", i64)
}


func (i64 *Int64) Scan(value any) error {
	if value == nil {
		*i64 = Int64(0)
		return nil
	}

	switch v := value.(type) {
	case int64:
		*i64 = Int64(v)
	case []byte, string:
		var str string
		if b, ok := v.([]byte); ok {
			str = string(b)
		} else {
			str = v.(string)
		}

		parsed, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into Int64: %w", v, err)
		}
		*i64 = Int64(parsed)
	default:
		return fmt.Errorf("cannot scan %T into Int64", value)
	}

	return nil
}


func (i64 Int64) Value() (driver.Value, error) {
	return int64(i64), nil
}


func NewInt64(input int64) Int64 {
	i64 := Int64(input)
	return i64
}
