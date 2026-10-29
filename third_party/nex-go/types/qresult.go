package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

var errorMask = 1 << 31





type QResult uint32


func (r QResult) WriteTo(writable Writable) {
	writable.WriteUInt32LE(uint32(r))
}


func (r *QResult) ExtractFrom(readable Readable) error {
	code, err := readable.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read QResult code. %s", err.Error())
	}

	*r = QResult(code)
	return nil
}


func (r QResult) Copy() RVType {
	return NewQResult(uint32(r))
}


func (r QResult) Equals(o RVType) bool {
	if _, ok := o.(QResult); !ok {
		return false
	}

	return r == o.(QResult)
}



func (r QResult) CopyRef() RVTypePtr {
	copied := r.Copy().(QResult)
	return &copied
}




func (r *QResult) Deref() RVType {
	return *r
}


func (r QResult) IsSuccess() bool {
	return int(r)&errorMask == 0
}


func (r QResult) IsError() bool {
	return int(r)&errorMask != 0
}


func (r QResult) String() string {
	return r.FormatToString(0)
}


func (r QResult) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("QResult{\n")

	if r.IsSuccess() {
		b.WriteString(fmt.Sprintf("%scode: %d (success)\n", indentationValues, r))
	} else {
		b.WriteString(fmt.Sprintf("%scode: %d (error)\n", indentationValues, r))
	}

	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}


func (r *QResult) Scan(value any) error {
	if value == nil {
		*r = QResult(0)
		return nil
	}

	switch v := value.(type) {
	case int64:
		*r = QResult(v)
	case string:
		parsed, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into QResult: %w", v, err)
		}
		*r = QResult(parsed)
	default:
		return fmt.Errorf("cannot scan %T into QResult", value)
	}

	return nil
}


func (r QResult) Value() (driver.Value, error) {
	return int64(r), nil
}


func NewQResult(input uint32) QResult {
	r := QResult(input)
	return r
}


func NewQResultSuccess(code uint32) QResult {
	return NewQResult(uint32(int(code) & ^errorMask))
}


func NewQResultError(code uint32) QResult {
	return NewQResult(uint32(int(code) | errorMask))
}
