

package encryption


type Algorithm interface {
	Key() []byte
	SetKey(key []byte) error
	Encrypt(payload []byte) ([]byte, error)
	Decrypt(payload []byte) ([]byte, error)
	Copy() Algorithm
}
