package types



type DataInterface interface {
	HoldableObject
	DataObjectID() RVType
}


type DataHolder = AnyObjectHolder[DataInterface]


func NewDataHolder() DataHolder {
	return DataHolder{}
}
