package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)






type List[T RVType] []T


func (l List[T]) WriteTo(writable Writable) {
	writable.WriteUInt32LE(uint32(len(l)))

	for _, v := range l {
		v.WriteTo(writable)
	}
}

func (l List[T]) extractType(t any, readable Readable) error {


	if ptr, ok := t.(RVTypePtr); ok {
		return ptr.ExtractFrom(readable)
	}



	return fmt.Errorf("Unsupported List type %T", t)
}


func (l *List[T]) ExtractFrom(readable Readable) error {
	length, err := readable.ReadUInt32LE()
	if err != nil {
		return err
	}

	slice := make([]T, 0, length)

	for i := 0; i < int(length); i++ {
		var value T
		if err := l.extractType(&value, readable); err != nil {
			return err
		}

		slice = append(slice, value)
	}

	*l = slice

	return nil
}


func (l List[T]) Copy() RVType {
	copied := make(List[T], 0)

	for _, v := range l {
		copied = append(copied, v.Copy().(T))
	}

	return copied
}


func (l List[T]) Equals(o RVType) bool {
	if _, ok := o.(List[T]); !ok {
		return false
	}

	other := o.(List[T])

	if len(l) != len(other) {
		return false
	}

	for i := 0; i < len(l); i++ {
		if !l[i].Equals(other[i]) {
			return false
		}
	}

	return true
}



func (l List[T]) CopyRef() RVTypePtr {
	copied := l.Copy().(List[T])
	return &copied
}




func (l *List[T]) Deref() RVType {
	return *l
}


func (l List[T]) Contains(checkValue T) bool {
	for _, v := range l {
		if v.Equals(checkValue) {
			return true
		}
	}

	return false
}


func (l List[T]) String() string {
	return fmt.Sprintf("%v", ([]T)(l))
}


func (l *List[T]) Scan(value any) error {
	if value == nil {
		*l = List[T]{}
		return nil
	}

	var pgArray string
	switch v := value.(type) {
	case []byte:
		pgArray = string(v)
	case string:
		pgArray = v
	default:
		return fmt.Errorf("unsupported Scan type for List: %T", value)
	}

	if len(pgArray) < 2 || pgArray[0] != '{' || pgArray[len(pgArray)-1] != '}' {
		return fmt.Errorf("invalid PostgreSQL array format: %s", pgArray)
	}

	pgArray = pgArray[1 : len(pgArray)-1]

	if pgArray == "" {
		*l = List[T]{}
		return nil
	}

	var elements []string

	if strings.HasPrefix(pgArray, "\"") {
		reader := csv.NewReader(strings.NewReader(pgArray))
		reader.Comma = ','

		var err error
		elements, err = reader.Read()

		if err != nil {
			return fmt.Errorf("failed to parse array elements: %w", err)
		}
	} else {
		elements = strings.Split(pgArray, ",")
	}

	var zero T
	isByteaType := false

	switch any(zero).(type) {
	case Buffer, QBuffer:
		isByteaType = true
	}

	result := make(List[T], 0, len(elements))

	for _, element := range elements {
		var new T
		ptr := any(&new)

		if scanner, ok := ptr.(sql.Scanner); ok {
			var scanValue any = element

			if isByteaType {
				hexStr := element
				hexStr = strings.TrimPrefix(hexStr, "\\\\x")
				hexStr = strings.TrimPrefix(hexStr, "\\x")
				bytes, err := hex.DecodeString(hexStr)

				if err != nil {
					return fmt.Errorf("failed to decode hex for bytea element %q: %w", element, err)
				}

				scanValue = bytes
			}

			if err := scanner.Scan(scanValue); err != nil {
				return fmt.Errorf("failed to scan element: %w", err)
			}

			result = append(result, new)
		} else {
			return fmt.Errorf("list element type %T does not implement sql.Scanner", zero)
		}
	}

	*l = result

	return nil
}


func (l List[T]) Value() (driver.Value, error) {
	if len(l) == 0 {
		return "{}", nil
	}

	var b strings.Builder
	b.WriteString("{")

	for i, new := range l {
		if i > 0 {
			b.WriteString(",")
		}

		if valuer, ok := any(new).(driver.Valuer); ok {
			v, err := valuer.Value()
			if err != nil {
				return nil, fmt.Errorf("failed to get value for element %d: %w", i, err)
			}

			switch val := v.(type) {
			case string:
				escaped := strings.ReplaceAll(val, "\\", "\\\\")
				escaped = strings.ReplaceAll(escaped, "\"", "\\\"")

				b.WriteString("\"")
				b.WriteString(escaped)
				b.WriteString("\"")
			case int64, float64, bool:
				b.WriteString(fmt.Sprintf("%v", val))
			case []byte:
				b.WriteString("\"\\\\x")
				b.WriteString(hex.EncodeToString(val))
				b.WriteString("\"")
			case time.Time:
				b.WriteString("\"")
				b.WriteString(val.Format("2006-01-02 15:04:05"))
				b.WriteString("\"")
			case nil:
				b.WriteString("NULL")
			default:
				return nil, fmt.Errorf("unsupported value type %T for element %d", val, i)
			}
		} else {
			return nil, fmt.Errorf("element type %T does not implement driver.Valuer", new)
		}
	}

	b.WriteString("}")

	return b.String(), nil
}


func NewList[T RVType]() List[T] {
	return make(List[T], 0)
}
