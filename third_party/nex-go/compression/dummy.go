package compression


type Dummy struct{}


func (d *Dummy) Compress(payload []byte) ([]byte, error) {
	return payload, nil
}


func (d *Dummy) Decompress(payload []byte) ([]byte, error) {
	return payload, nil
}


func (d *Dummy) Copy() Algorithm {
	return NewDummyCompression()
}


func NewDummyCompression() *Dummy {
	return &Dummy{}
}
