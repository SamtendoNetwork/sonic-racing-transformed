package types

import (
	"fmt"
	"strings"
)



type ClassVersionContainer struct {
	Structure
	ClassVersions Map[String, UInt16] `json:"class_versions" db:"class_versions" bson:"class_versions" xml:"ClassVersions"`
}


func (cvc ClassVersionContainer) WriteTo(writable Writable) {
	cvc.ClassVersions.WriteTo(writable)
}


func (cvc *ClassVersionContainer) ExtractFrom(readable Readable) error {
	return cvc.ClassVersions.ExtractFrom(readable)
}


func (cvc ClassVersionContainer) Copy() RVType {
	copied := NewClassVersionContainer()
	copied.ClassVersions = cvc.ClassVersions.Copy().(Map[String, UInt16])

	return copied
}


func (cvc ClassVersionContainer) Equals(o RVType) bool {
	if _, ok := o.(ClassVersionContainer); !ok {
		return false
	}

	return cvc.ClassVersions.Equals(o)
}



func (cvc ClassVersionContainer) CopyRef() RVTypePtr {
	copied := cvc.Copy().(ClassVersionContainer)
	return &copied
}




func (cvc *ClassVersionContainer) Deref() RVType {
	return *cvc
}


func (cvc ClassVersionContainer) String() string {
	return cvc.FormatToString(0)
}


func (cvc ClassVersionContainer) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("ClassVersionContainer{\n")
	b.WriteString(fmt.Sprintf("%sStructureVersion: %d,\n", indentationValues, cvc.StructureVersion))
	b.WriteString(fmt.Sprintf("%sClassVersions: %s\n", indentationValues, cvc.ClassVersions))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}


func NewClassVersionContainer() ClassVersionContainer {
	cvc := ClassVersionContainer{
		ClassVersions: NewMap[String, UInt16](),
	}

	return cvc
}
