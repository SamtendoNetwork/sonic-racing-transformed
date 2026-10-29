package nex


type ByteStreamSettings struct {
	StringLengthSize   int
	PIDSize            int
	UseStructureHeader bool
}


func NewByteStreamSettings() *ByteStreamSettings {
	return &ByteStreamSettings{
		StringLengthSize:   2,
		PIDSize:            4,
		UseStructureHeader: false,
	}
}
