package encryption

import (
	"crypto/rc4"
)


type RC4 struct {
	key             []byte
	cipher          *rc4.Cipher
	decipher        *rc4.Cipher
	cipheredCount   uint64
	decipheredCount uint64
}


func (r *RC4) Key() []byte {
	return r.key
}


func (r *RC4) SetKey(key []byte) error {
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


func (r *RC4) Encrypt(payload []byte) ([]byte, error) {
	ciphered := make([]byte, len(payload))

	r.cipher.XORKeyStream(ciphered, payload)

	r.cipheredCount += uint64(len(payload))

	return ciphered, nil
}


func (r *RC4) Decrypt(payload []byte) ([]byte, error) {
	deciphered := make([]byte, len(payload))

	r.decipher.XORKeyStream(deciphered, payload)

	r.decipheredCount += uint64(len(payload))

	return deciphered, nil
}


func (r *RC4) Copy() Algorithm {
	copied := NewRC4Encryption()

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


func NewRC4Encryption() *RC4 {
	encryption := &RC4{
		key: make([]byte, 0),
	}

	encryption.SetKey([]byte("CD&ML"))

	return encryption
}
