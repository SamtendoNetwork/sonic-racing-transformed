

package compression


type Algorithm interface {
	Compress(payload []byte) ([]byte, error)
	Decompress(payload []byte) ([]byte, error)
	Copy() Algorithm
}
