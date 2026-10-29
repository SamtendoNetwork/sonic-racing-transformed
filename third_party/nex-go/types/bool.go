package types

import (
	"database/sql/driver"
	"fmt"
)


type Bool bool


func (b Bool) WriteTo(writable Writable) {
	writable.WriteBool(bool(b))
}


func (b *Bool) ExtractFrom(readable Readable) error {
	value, err := readable.ReadBool()
	if err != nil {
		return err
	}

	*b = Bool(value)
	return nil
}


func (b Bool) Copy() RVType {
	return NewBool(bool(b))
}


func (b Bool) Equals(o RVType) bool {
	other, ok := o.(Bool)
	if !ok {
		return false
	}

	return b == other
}



func (b Bool) CopyRef() RVTypePtr {
	copied := b.Copy().(Bool)
	return &copied
}




func (b *Bool) Deref() RVType {
	return *b
}


func (b Bool) String() string {
	return fmt.Sprintf("%t", b)
}


func (b *Bool) Scan(value any) error {
	if value == nil {
		*b = Bool(false)
		return nil
	}

	switch v := value.(type) {
	case bool:
		*b = Bool(v)
	case string:
		*b = Bool(v == "t" || v == "true")
	default:
		return fmt.Errorf("cannot scan %T into Bool", value)
	}

	return nil
}


func (b Bool) Value() (driver.Value, error) {
	return bool(b), nil
}


func NewBool(input bool) Bool {
	b := Bool(input)
	return b
}
