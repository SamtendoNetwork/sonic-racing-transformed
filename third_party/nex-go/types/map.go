package types

import (
	"fmt"
	"strings"
)








type Map[K comparable, V RVType] map[K]V

func (m Map[K, V]) writeType(t any, writable Writable) {


	if rvt, ok := t.(interface{ WriteTo(writable Writable) }); ok {
		rvt.WriteTo(writable)
	}
}


func (m Map[K, V]) WriteTo(writable Writable) {
	writable.WriteUInt32LE(uint32(len(m)))

	for key, value := range m {
		m.writeType(key, writable)
		m.writeType(value, writable)
	}
}

func (m Map[K, V]) extractType(t any, readable Readable) error {


	if ptr, ok := t.(RVTypePtr); ok {
		return ptr.ExtractFrom(readable)
	}



	return fmt.Errorf("Unsupported Map type %T", t)
}


func (m *Map[K, V]) ExtractFrom(readable Readable) error {
	length, err := readable.ReadUInt32LE()
	if err != nil {
		return err
	}

	extracted := make(Map[K, V])

	for i := 0; i < int(length); i++ {
		var key K
		if err := m.extractType(&key, readable); err != nil {
			return err
		}

		var value V
		if err := m.extractType(&value, readable); err != nil {
			return err
		}

		extracted[key] = value
	}

	*m = extracted

	return nil
}

func (m Map[K, V]) copyType(t any) RVType {


	if rvt, ok := t.(RVType); ok {
		return rvt.Copy()
	}


	return nil
}


func (m Map[K, V]) Copy() RVType {
	copied := make(Map[K, V])

	for key, value := range m {
		copied[m.copyType(key).(K)] = value.Copy().(V)
	}

	return copied
}

func (m Map[K, V]) typesEqual(t1, t2 any) bool {


	if rvt1, ok := t1.(RVType); ok {
		if rvt2, ok := t2.(RVType); ok {
			return rvt1.Equals(rvt2)
		}
	}

	return false
}


func (m Map[K, V]) Equals(o RVType) bool {
	if _, ok := o.(Map[K, V]); !ok {
		return false
	}

	other := o.(Map[K, V])

	if len(m) != len(other) {
		return false
	}

	for key, value := range m {
		if otherValue, ok := other[key]; !ok || m.typesEqual(&value, &otherValue) {
			return false
		}
	}

	return true
}



func (m Map[K, V]) CopyRef() RVTypePtr {
	copied := m.Copy().(Map[K, V])
	return &copied
}




func (m *Map[K, V]) Deref() RVType {
	return *m
}


func (m Map[K, V]) String() string {
	return m.FormatToString(0)
}


func (m Map[K, V]) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	if len(m) == 0 {
		b.WriteString("{}\n")
	} else {
		b.WriteString("{\n")

		for key, value := range m {

			b.WriteString(fmt.Sprintf("%s%v: %v,\n", indentationValues, key, value))
		}

		b.WriteString(fmt.Sprintf("%s}\n", indentationEnd))
	}

	return b.String()
}


func NewMap[K comparable, V RVType]() Map[K, V] {
	return make(Map[K, V])
}
