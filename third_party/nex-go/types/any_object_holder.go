package types

import (
	"fmt"
	"strings"
)


type HoldableObject interface {
	RVType
	ObjectID() RVType
}


var AnyObjectHolderObjects = make(map[RVType]HoldableObject)


func RegisterObjectHolderType(rvType HoldableObject) {
	AnyObjectHolderObjects[rvType.ObjectID()] = rvType
}


type AnyObjectHolder[T HoldableObject] struct {
	Object T
}


func (aoh AnyObjectHolder[T]) WriteTo(writable Writable) {
	contentWritable := writable.CopyNew()

	aoh.Object.WriteTo(contentWritable)

	objectBuffer := NewBuffer(contentWritable.Bytes())

	objectBufferLength := uint32(len(objectBuffer) + 4)

	aoh.Object.ObjectID().WriteTo(writable)
	writable.WriteUInt32LE(objectBufferLength)
	objectBuffer.WriteTo(writable)
}


func (aoh *AnyObjectHolder[T]) ExtractFrom(readable Readable) error {
	var err error


	identifier := NewString("")
	err = identifier.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read AnyObjectHolder identifier. %s", err.Error())
	}

	length := NewUInt32(0)
	err = length.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read AnyObjectHolder length. %s", err.Error())
	}


	bufferLength := NewUInt32(0)
	err = bufferLength.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read AnyObjectHolder buffer length. %s", err.Error())
	}

	if _, ok := AnyObjectHolderObjects[identifier]; !ok {
		return fmt.Errorf("Unknown AnyObjectHolder identifier: %s", identifier)
	}

	ptr := AnyObjectHolderObjects[identifier].CopyRef()

	if err := ptr.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read AnyObjectHolder object. %s", err.Error())
	}

	var ok bool
	if aoh.Object, ok = ptr.Deref().(T); !ok {
		return fmt.Errorf("Input AnyObjectHolder object %s is invalid", identifier)
	}

	return nil
}


func (aoh AnyObjectHolder[T]) Copy() RVType {
	copied := NewAnyObjectHolder[T]()

	copied.Object = aoh.Object.Copy().(T)

	return copied
}


func (aoh AnyObjectHolder[T]) Equals(o RVType) bool {
	if _, ok := o.(AnyObjectHolder[T]); !ok {
		return false
	}

	other := o.(AnyObjectHolder[T])

	return aoh.Object.Equals(other.Object)
}



func (aoh AnyObjectHolder[T]) CopyRef() RVTypePtr {
	copied := aoh.Copy().(AnyObjectHolder[T])
	return &copied
}




func (aoh *AnyObjectHolder[T]) Deref() RVType {
	return *aoh
}


func (aoh AnyObjectHolder[T]) String() string {
	return aoh.FormatToString(0)
}


func (aoh AnyObjectHolder[T]) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("AnyDataHolder{\n")
	fmt.Fprintf(&b, "%sIdentifier: %v,\n", indentationValues, aoh.Object.ObjectID())
	fmt.Fprintf(&b, "%sObject: %v\n", indentationValues, aoh.Object)
	fmt.Fprintf(&b, "%s}", indentationEnd)

	return b.String()
}


func NewAnyObjectHolder[T HoldableObject]() AnyObjectHolder[T] {
	return AnyObjectHolder[T]{}
}
