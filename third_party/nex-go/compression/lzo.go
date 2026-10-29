package compression

import (
	"bytes"
	"fmt"

	"github.com/rasky/go-lzo"
)




type LZO struct{}


func (l *LZO) Compress(payload []byte) ([]byte, error) {
	compressed := lzo.Compress1X(payload)

	compressionRatio := len(payload)/len(compressed) + 1

	result := make([]byte, len(compressed)+1)

	result[0] = byte(compressionRatio)

	copy(result[1:], compressed)

	return result, nil
}


func (l *LZO) Decompress(payload []byte) ([]byte, error) {
	compressionRatio := payload[0]
	compressed := payload[1:]

	if compressionRatio == 0 {

		return compressed, nil
	}

	reader := bytes.NewReader(compressed)
	decompressed, err := lzo.Decompress1X(reader, len(compressed), 0)
	if err != nil {
		return []byte{}, err
	}

	ratioCheck := len(decompressed)/len(compressed) + 1

	if ratioCheck != int(compressionRatio) {
		return []byte{}, fmt.Errorf("Failed to decompress payload. Got bad ratio. Expected %d, got %d", compressionRatio, ratioCheck)
	}

	return decompressed, nil
}


func (l *LZO) Copy() Algorithm {
	return NewLZOCompression()
}


func NewLZOCompression() *LZO {
	return &LZO{}
}
