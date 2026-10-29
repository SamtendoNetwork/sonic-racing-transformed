package encryption


type Dummy struct {
	key []byte
}


func (d *Dummy) Key() []byte {
	return d.key
}


func (d *Dummy) SetKey(key []byte) error {
	d.key = key

	return nil
}


func (d *Dummy) Encrypt(payload []byte) ([]byte, error) {
	return payload, nil
}


func (d *Dummy) Decrypt(payload []byte) ([]byte, error) {
	return payload, nil
}


func (d *Dummy) Copy() Algorithm {
	copied := NewDummyEncryption()

	copied.key = d.key

	return copied
}


func NewDummyEncryption() *Dummy {
	return &Dummy{
		key: make([]byte, 0),
	}
}
