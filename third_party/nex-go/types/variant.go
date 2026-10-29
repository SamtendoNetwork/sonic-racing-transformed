package types

import (
	"fmt"
	"strings"
)


var VariantTypes = make(map[UInt8]RVType)


func RegisterVariantType(id UInt8, rvType RVType) {
	VariantTypes[id] = rvType
}



type Variant struct {
	TypeID UInt8  `json:"type_id" db:"type_id" bson:"type_id" xml:"TypeID"`
	Type   RVType `json:"type" db:"type" bson:"type" xml:"Type"`
}


func (v Variant) WriteTo(writable Writable) {
	v.TypeID.WriteTo(writable)

	if v.Type != nil {
		v.Type.WriteTo(writable)
	}
}


func (v *Variant) ExtractFrom(readable Readable) error {
	err := v.TypeID.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read Variant type ID. %s", err.Error())
	}

	typeID := v.TypeID


	if typeID == 0 {
		return nil
	}

	if _, ok := VariantTypes[typeID]; !ok {
		return fmt.Errorf("Invalid Variant type ID %d", typeID)
	}



	ptr := VariantTypes[typeID].CopyRef()
	if err := ptr.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read Variant type data. %s", err.Error())
	}

	v.Type = ptr.Deref()

	return nil
}


func (v Variant) Copy() RVType {
	copied := NewVariant()

	copied.TypeID = v.TypeID.Copy().(UInt8)

	if v.Type != nil {
		copied.Type = v.Type.Copy()
	}

	return copied
}


func (v Variant) Equals(o RVType) bool {
	if _, ok := o.(Variant); !ok {
		return false
	}

	other := o.(Variant)

	if !v.TypeID.Equals(other.TypeID) {
		return false
	}

	if v.Type != nil {
		return v.Type.Equals(other.Type)
	}

	return true
}



func (v Variant) CopyRef() RVTypePtr {
	copied := v.Copy().(Variant)
	return &copied
}




func (v *Variant) Deref() RVType {
	return *v
}


func (v Variant) String() string {
	return v.FormatToString(0)
}


func (v Variant) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("Variant{\n")
	b.WriteString(fmt.Sprintf("%sTypeID: %s,\n", indentationValues, v.TypeID))

	if v.Type != nil {
		b.WriteString(fmt.Sprintf("%sType: %s\n", indentationValues, v.Type))
	} else {
		b.WriteString(fmt.Sprintf("%sType: None\n", indentationValues))
	}

	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}


func NewVariant() Variant {

	return Variant{
		TypeID: NewUInt8(0),
		Type:   nil,
	}
}
