
package types



type RVType interface {
	WriteTo(writable Writable)
	Copy() RVType
	CopyRef() RVTypePtr
	Equals(other RVType) bool
}



type RVTypePtr interface {
	RVType
	ExtractFrom(readable Readable) error
	Deref() RVType
}
