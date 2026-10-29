package types

import (
	"fmt"
	"strings"
)



type Data struct {
	Structure
}


func (d Data) ObjectID() RVType {
	return d.DataObjectID()
}


func (d Data) DataObjectID() RVType {
	return NewString("Data")
}


func (d Data) WriteTo(writable Writable) {
	d.WriteHeaderTo(writable, 0)
}


func (d *Data) ExtractFrom(readable Readable) error {
	if err := d.ExtractHeaderFrom(readable); err != nil {
		return fmt.Errorf("Failed to read Data header. %s", err.Error())
	}

	return nil
}


func (d Data) Copy() RVType {
	copied := NewData()
	copied.StructureVersion = d.StructureVersion

	return copied
}


func (d Data) Equals(o RVType) bool {
	if _, ok := o.(Data); !ok {
		return false
	}

	other := o.(Data)

	return d.StructureVersion == other.StructureVersion
}



func (d Data) CopyRef() RVTypePtr {
	copied := d.Copy().(Data)
	return &copied
}




func (d *Data) Deref() RVType {
	return *d
}


func (d Data) String() string {
	return d.FormatToString(0)
}


func (d Data) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("Data{\n")
	b.WriteString(fmt.Sprintf("%sStructureVersion: %d\n", indentationValues, d.StructureVersion))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}


func NewData() Data {
	return Data{}
}
