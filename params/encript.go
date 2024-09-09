package params

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

// Creating 32 byte length hash using the password. This mean the password can be any length.
func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Encryption the data. The password can be of any complexity and length
func encrypt(dataRaw string, passphrase string) ([]byte, error) {
	data := []byte(dataRaw)

	// No need error, because using Hash of passphrase
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// Decryption of data that was previously encrypted using the same password
func decrypt(dataEnc string, passphrase string) ([]byte, error) {
	data := []byte(dataEnc)

	// No need error, because using Hash of passphrase
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
