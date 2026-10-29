package types

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)



type String string


func (s String) WriteTo(writable Writable) {
	s = s + "\x00"
	strLength := len(s)

	if writable.StringLengthSize() == 4 {
		writable.WriteUInt32LE(uint32(strLength))
	} else {
		writable.WriteUInt16LE(uint16(strLength))
	}

	writable.Write([]byte(s))
}


func (s *String) ExtractFrom(readable Readable) error {
	var length uint64
	var err error

	if readable.StringLengthSize() == 4 {
		l, e := readable.ReadUInt32LE()
		length = uint64(l)
		err = e
	} else {
		l, e := readable.ReadUInt16LE()
		length = uint64(l)
		err = e
	}

	if err != nil {
		return fmt.Errorf("Failed to read NEX string length. %s", err.Error())
	}

	if readable.Remaining() < length {
		return errors.New("NEX string length longer than data size")
	}

	stringData, err := readable.Read(length)
	if err != nil {
		return fmt.Errorf("Failed to read NEX string length. %s", err.Error())
	}

	str := strings.TrimRight(string(stringData), "\x00")

	*s = String(str)
	return nil
}


func (s String) Copy() RVType {
	return NewString(string(s))
}


func (s String) Equals(o RVType) bool {
	if _, ok := o.(String); !ok {
		return false
	}

	return s == o.(String)
}



func (s String) CopyRef() RVTypePtr {
	copied := s.Copy().(String)
	return &copied
}




func (s *String) Deref() RVType {
	return *s
}


func (s String) String() string {
	return fmt.Sprintf("%q", string(s))
}


func (s *String) Scan(value any) error {
	if value == nil {
		*s = String("")
		return nil
	}

	if _, ok := value.(string); !ok {
		return fmt.Errorf("cannot scan %T into String", value)
	}

	*s = String(value.(string))

	return nil
}


func (s String) Value() (driver.Value, error) {
	return string(s), nil
}


func NewString(input string) String {
	s := String(input)
	return s
}
