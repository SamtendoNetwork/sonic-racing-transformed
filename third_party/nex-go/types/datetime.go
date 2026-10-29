package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"
)




type DateTime uint64


func (dt DateTime) WriteTo(writable Writable) {
	writable.WriteUInt64LE(uint64(dt))
}


func (dt *DateTime) ExtractFrom(readable Readable) error {
	value, err := readable.ReadUInt64LE()
	if err != nil {
		return fmt.Errorf("Failed to read DateTime value. %s", err.Error())
	}

	*dt = DateTime(value)
	return nil
}


func (dt DateTime) Copy() RVType {
	return NewDateTime(uint64(dt))
}


func (dt DateTime) Equals(o RVType) bool {
	if _, ok := o.(DateTime); !ok {
		return false
	}

	return dt == o.(DateTime)
}



func (dt DateTime) CopyRef() RVTypePtr {
	copied := dt.Copy().(DateTime)
	return &copied
}




func (dt *DateTime) Deref() RVType {
	return *dt
}


func (dt *DateTime) Make(year, month, day, hour, minute, second int) DateTime {
	*dt = DateTime(second | (minute << 6) | (hour << 12) | (day << 17) | (month << 22) | (year << 26))

	return *dt
}


func (dt *DateTime) FromTimestamp(timestamp time.Time) DateTime {
	year := timestamp.Year()
	month := int(timestamp.Month())
	day := timestamp.Day()
	hour := timestamp.Hour()
	minute := timestamp.Minute()
	second := timestamp.Second()

	return dt.Make(year, month, day, hour, minute, second)
}


func (dt DateTime) Now() DateTime {
	return dt.FromTimestamp(time.Now().UTC())
}


func (dt DateTime) Second() int {
	return int(dt & 63)
}


func (dt DateTime) Minute() int {
	return int((dt >> 6) & 63)
}


func (dt DateTime) Hour() int {
	return int((dt >> 12) & 31)
}


func (dt DateTime) Day() int {
	return int((dt >> 17) & 31)
}


func (dt DateTime) Month() time.Month {
	return time.Month((dt >> 22) & 15)
}


func (dt DateTime) Year() int {
	return int(dt >> 26)
}


func (dt DateTime) Standard() time.Time {
	return time.Date(
		dt.Year(),
		dt.Month(),
		dt.Day(),
		dt.Hour(),
		dt.Minute(),
		dt.Second(),
		0,
		time.UTC,
	)
}


func (dt DateTime) String() string {
	return dt.FormatToString(0)
}


func (dt DateTime) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("DateTime{\n")
	b.WriteString(fmt.Sprintf("%svalue: %d (%s)\n", indentationValues, dt, dt.Standard().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}









func (dt DateTime) Value() (driver.Value, error) {

	if dt == 0 {
		return nil, nil
	}


	return dt.Standard(), nil
}








func (dt *DateTime) Scan(value any) error {

	if value == nil {
		*dt = 0
		return nil
	}




	switch v := value.(type) {
	case int64:
		*dt = DateTime(uint64(v))
		return nil
	case uint64:
		*dt = DateTime(v)
		return nil
	case []byte:
		return dt.scanSQLString(string(v))
	case string:
		return dt.scanSQLString(v)
	case time.Time:
		dt.FromTimestamp(v)
		return nil
	default:
		return fmt.Errorf("DateTime.Scan: cannot convert %T to DateTime", value)
	}
}




func (dt *DateTime) scanSQLString(str string) error {



	timeFormats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05Z07:00",
	}

	for _, format := range timeFormats {
		if t, err := time.Parse(format, str); err == nil {
			dt.FromTimestamp(t)
			return nil
		}
	}





	if val, err := strconv.ParseUint(str, 10, 64); err == nil {
		*dt = DateTime(val)
		return nil
	}


	return fmt.Errorf("DateTime.Scan: cannot convert %q to DateTime", str)
}


func NewDateTime(input uint64) DateTime {
	dt := DateTime(input)
	return dt
}
