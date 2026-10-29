package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)


type Float float32


func (f Float) WriteTo(writable Writable) {
	writable.WriteFloat32LE(float32(f))
}


func (f *Float) ExtractFrom(readable Readable) error {
	value, err := readable.ReadFloat32LE()
	if err != nil {
		return err
	}

	*f = Float(value)
	return nil
}


func (f Float) Copy() RVType {
	return NewFloat(float32(f))
}


func (f Float) Equals(o RVType) bool {
	other, ok := o.(Float)
	if !ok {
		return false
	}

	return f == other
}



func (f Float) CopyRef() RVTypePtr {
	copied := f.Copy().(Float)
	return &copied
}




func (f *Float) Deref() RVType {
	return *f
}


func (f Float) String() string {
	return fmt.Sprintf("%f", f)
}


func (f *Float) Scan(value any) error {
	if value == nil {
		*f = Float(0.0)
		return nil
	}

	switch v := value.(type) {
	case float64:
		*f = Float(value.(float64))
	case string:
		parsed, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into Float: %w", v, err)
		}
		*f = Float(parsed)
	default:
		return fmt.Errorf("cannot scan %T into Float", value)
	}

	return nil
}


func (f Float) Value() (driver.Value, error) {
	return float64(f), nil
}


func NewFloat(input float32) Float {
	f := Float(input)
	return f
}
