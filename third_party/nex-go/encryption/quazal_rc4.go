package encryption

import (
	"crypto/rc4"
)


type QuazalRC4 struct {
	key             []byte
	cipher          *rc4.Cipher
	decipher        *rc4.Cipher
	cipheredCount   uint64
	decipheredCount uint64
}


func (r *QuazalRC4) Key() []byte {
	return r.key
}


func (r *QuazalRC4) SetKey(key []byte) error {
	r.key = key

	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return err
	}

	decipher, err := rc4.NewCipher(key)
	if err != nil {
		return err
	}

	r.cipher = cipher
	r.decipher = decipher

	return nil
}


func (r *QuazalRC4) Encrypt(payload []byte) ([]byte, error) {
	r.SetKey([]byte("CD&ML"))

	ciphered := make([]byte, len(payload))

	r.cipher.XORKeyStream(ciphered, payload)

	r.cipheredCount += uint64(len(payload))

	return ciphered, nil
}


func (r *QuazalRC4) Decrypt(payload []byte) ([]byte, error) {
	r.SetKey([]byte("CD&ML"))

	deciphered := make([]byte, len(payload))

	r.decipher.XORKeyStream(deciphered, payload)

	r.decipheredCount += uint64(len(payload))

	return deciphered, nil
}


func (r *QuazalRC4) Copy() Algorithm {
	copied := NewQuazalRC4Encryption()

	copied.SetKey(r.key)




	for i := 0; i < int(r.cipheredCount); i++ {
		copied.cipher.XORKeyStream([]byte{0}, []byte{0})
	}

	for i := 0; i < int(r.decipheredCount); i++ {
		copied.decipher.XORKeyStream([]byte{0}, []byte{0})
	}

	copied.cipheredCount = r.cipheredCount
	copied.decipheredCount = r.decipheredCount

	return copied
}


func NewQuazalRC4Encryption() *QuazalRC4 {
	return &QuazalRC4{
		key: make([]byte, 0),
	}
}
