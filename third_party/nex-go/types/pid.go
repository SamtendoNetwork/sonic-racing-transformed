package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)





type PID uint64


func (p PID) WriteTo(writable Writable) {
	if writable.PIDSize() == 8 {
		writable.WriteUInt64LE(uint64(p))
	} else {
		writable.WriteUInt32LE(uint32(p))
	}
}


func (p *PID) ExtractFrom(readable Readable) error {
	var pid uint64
	var err error

	if readable.PIDSize() == 8 {
		pid, err = readable.ReadUInt64LE()
	} else {
		p, e := readable.ReadUInt32LE()

		pid = uint64(p)
		err = e
	}

	if err != nil {
		return err
	}

	*p = PID(pid)
	return nil
}


func (p PID) Copy() RVType {
	return NewPID(uint64(p))
}


func (p PID) Equals(o RVType) bool {
	if _, ok := o.(PID); !ok {
		return false
	}

	return p == o.(PID)
}



func (p PID) CopyRef() RVTypePtr {
	copied := p.Copy().(PID)
	return &copied
}




func (p *PID) Deref() RVType {
	return *p
}


func (p PID) String() string {
	return p.FormatToString(0)
}


func (p PID) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("PID{\n")
	b.WriteString(fmt.Sprintf("%spid: %d\n", indentationValues, p))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}


func (p *PID) Scan(value any) error {
	if value == nil {
		*p = PID(0)
		return nil
	}

	switch v := value.(type) {
	case int64:
		*p = PID(uint64(v))
	case []byte, string:
		var str string
		if b, ok := v.([]byte); ok {
			str = string(b)
		} else {
			str = v.(string)
		}

		parsed, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into PID: %w", v, err)
		}
		*p = PID(parsed)
	default:
		return fmt.Errorf("cannot scan %T into PID", value)
	}

	return nil
}


func (p PID) Value() (driver.Value, error) {
	return fmt.Sprintf("%d", uint64(p)), nil
}


func NewPID(input uint64) PID {
	p := PID(input)
	return p
}
