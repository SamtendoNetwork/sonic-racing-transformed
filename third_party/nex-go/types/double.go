package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)


type Double float64


func (d Double) WriteTo(writable Writable) {
	writable.WriteFloat64LE(float64(d))
}


func (d *Double) ExtractFrom(readable Readable) error {
	value, err := readable.ReadFloat64LE()
	if err != nil {
		return err
	}

	*d = Double(value)
	return nil
}


func (d Double) Copy() RVType {
	return NewDouble(float64(d))
}


func (d Double) Equals(o RVType) bool {
	other, ok := o.(Double)
	if !ok {
		return false
	}

	return d == other
}



func (d Double) CopyRef() RVTypePtr {
	copied := d.Copy().(Double)
	return &copied
}




func (d *Double) Deref() RVType {
	return *d
}


func (d Double) String() string {
	return fmt.Sprintf("%f", d)
}


func (d *Double) Scan(value any) error {
	if value == nil {
		*d = Double(0.0)
		return nil
	}

	switch v := value.(type) {
	case float64:
		*d = Double(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into Double: %w", v, err)
		}
		*d = Double(parsed)
	default:
		return fmt.Errorf("cannot scan %T into Double", value)
	}

	return nil
}


func (d Double) Value() (driver.Value, error) {
	return float64(d), nil
}


func NewDouble(input float64) Double {
	d := Double(input)
	return d
}
